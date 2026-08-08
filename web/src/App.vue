<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  AlertCircle,
  CheckCircle2,
  Copy,
  Folder,
  Inbox,
  Loader2,
  Mail,
  Plus,
  RefreshCw,
  ShieldCheck,
} from "lucide-vue-next";

type ApiResponse<T> = {
  success: boolean;
  message?: string;
  data?: T;
};

type Account = {
  id: string;
  name: string;
  real_email?: string;
  icloud_email?: string;
  host?: string;
  status?: string;
  alias_total?: number;
  alias_active?: number;
  last_validated?: string;
};

type Alias = {
  email: string;
  anonymousId: string;
  label: string;
  active: boolean;
  createdAt?: string;
};

type Message = {
  id: string;
  uid?: string;
  folder?: string;
  from: string;
  to: string;
  subject: string;
  date: string;
  preview: string;
};

type FolderOption = {
  name: string;
  role: string;
};

type AliasesData = {
  account_id: string;
  count: number;
  aliases: Alias[] | null;
};

type MailboxesData = {
  account_id: string;
  folders: FolderOption[] | null;
};

type InboxData = {
  account_id: string;
  alias?: string;
  folder?: string;
  count: number;
  method: string;
  messages: Message[] | null;
};

type CreateData = {
  email: string;
  label: string;
  created_at: string;
  account_id: string;
};

const defaultFolders: FolderOption[] = [
  { name: "all", role: "all" },
  { name: "INBOX", role: "inbox" },
  { name: "Junk", role: "junk" },
];

const accounts = ref<Account[]>([]);
const aliases = ref<Alias[]>([]);
const folders = ref<FolderOption[]>(defaultFolders);
const messages = ref<Message[]>([]);
const selectedAccountId = ref("");
const selectedAlias = ref("");
const selectedFolder = ref("all");
const newLabel = ref("");
const notice = ref("");
const error = ref("");
const busy = ref({
  accounts: false,
  aliases: false,
  folders: false,
  create: false,
  inbox: false,
});
const inboxMeta = ref({ method: "", count: 0, folder: "all" });

const activeAccount = computed(() =>
  accounts.value.find((account) => account.id === selectedAccountId.value),
);
const activeAliases = computed(() => aliases.value.filter((item) => item.active).length);
const inactiveAliases = computed(() => aliases.value.length - activeAliases.value);
const selectedAliasInfo = computed(() =>
  aliases.value.find((item) => item.email === selectedAlias.value),
);
const folderOptions = computed(() => {
  const byName = new Map<string, FolderOption>();
  for (const item of defaultFolders) byName.set(item.name, item);
  for (const item of folders.value) {
    if (item.role === "inbox" || item.role === "junk") {
      byName.set(item.name, item);
    }
  }
  return Array.from(byName.values());
});

async function api<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
    ...init,
  });
  const body = (await response.json()) as ApiResponse<T>;
  if (!response.ok || !body.success) {
    throw new Error(body.message || `请求失败: ${response.status}`);
  }
  return body.data as T;
}

function setError(err: unknown) {
  error.value = err instanceof Error ? err.message : String(err);
}

function clearFeedback() {
  error.value = "";
  notice.value = "";
}

async function loadAccounts() {
  busy.value.accounts = true;
  clearFeedback();
  try {
    const data = await api<Account[]>("/api/accounts");
    accounts.value = data;
    if (!selectedAccountId.value && data.length > 0) {
      selectedAccountId.value = data[0].id;
    }
  } catch (err) {
    setError(err);
  } finally {
    busy.value.accounts = false;
  }
}

async function loadMailboxes() {
  if (!selectedAccountId.value) return;
  busy.value.folders = true;
  try {
    const data = await api<MailboxesData>(`/api/mailboxes?account_id=${selectedAccountId.value}`);
    folders.value = data.folders?.length ? data.folders : defaultFolders;
  } catch {
    folders.value = defaultFolders;
  } finally {
    busy.value.folders = false;
  }
}

async function loadAliases() {
  if (!selectedAccountId.value) return;
  busy.value.aliases = true;
  clearFeedback();
  try {
    const data = await api<AliasesData>(`/api/aliases?account_id=${selectedAccountId.value}`);
    aliases.value = data.aliases ?? [];
    if (selectedAlias.value && !aliases.value.some((item) => item.email === selectedAlias.value)) {
      selectedAlias.value = "";
    }
  } catch (err) {
    setError(err);
  } finally {
    busy.value.aliases = false;
  }
}

async function createAlias() {
  if (!selectedAccountId.value) return;
  busy.value.create = true;
  clearFeedback();
  try {
    const label = newLabel.value.trim() || `Web 管理台 ${new Date().toLocaleString()}`;
    const created = await api<CreateData>("/api/create", {
      method: "POST",
      body: JSON.stringify({ account_id: selectedAccountId.value, label }),
    });
    notice.value = `已创建 ${created.email}`;
    selectedAlias.value = created.email;
    selectedFolder.value = "all";
    newLabel.value = "";
    await loadAliases();
    await loadInbox(created.email);
  } catch (err) {
    setError(err);
  } finally {
    busy.value.create = false;
  }
}

async function loadInbox(alias = selectedAlias.value) {
  if (!selectedAccountId.value) return;
  busy.value.inbox = true;
  clearFeedback();
  try {
    const params = new URLSearchParams({
      account_id: selectedAccountId.value,
      limit: "50",
      days: "30",
      folder: selectedFolder.value,
    });
    if (alias) params.set("alias", alias);
    const data = await api<InboxData>(`/api/inbox?${params.toString()}`);
    messages.value = data.messages ?? [];
    inboxMeta.value = {
      method: data.method || "unknown",
      count: data.count,
      folder: data.folder || selectedFolder.value,
    };
  } catch (err) {
    setError(err);
  } finally {
    busy.value.inbox = false;
  }
}

async function refreshAll() {
  await loadAccounts();
  await loadMailboxes();
  await loadAliases();
  await loadInbox(selectedAlias.value);
}

async function handleAccountChange() {
  selectedAlias.value = "";
  await loadMailboxes();
  await loadAliases();
  await loadInbox();
}

async function chooseAlias(alias: Alias) {
  selectedAlias.value = alias.email;
  selectedFolder.value = "all";
  await loadInbox(alias.email);
}

async function copyText(text: string) {
  await navigator.clipboard.writeText(text);
  notice.value = "已复制到剪贴板";
}

function clearAliasSelection() {
  selectedAlias.value = "";
  void loadInbox();
}

function formatDate(value?: string) {
  if (!value) return "未知";
  const asNumber = Number(value);
  const date = Number.isFinite(asNumber) && value.length > 10 ? new Date(asNumber) : new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

function folderLabel(optionOrName?: FolderOption | string) {
  const option =
    typeof optionOrName === "string"
      ? folderOptions.value.find((item) => item.name === optionOrName || item.role === optionOrName)
      : optionOrName;
  const role = option?.role || optionOrName;
  if (role === "all") return "全部";
  if (role === "inbox") return "收件箱";
  if (role === "junk") return "垃圾邮件";
  return option?.name || String(optionOrName || "未知");
}

onMounted(async () => {
  await loadAccounts();
  await loadMailboxes();
  await loadAliases();
  await loadInbox();
});
</script>

<template>
  <main class="app-shell">
    <header class="topbar">
      <div>
        <p class="product-label">iCloud Hide My Email</p>
        <h1>隐私邮箱管理台</h1>
      </div>
      <button class="icon-button" type="button" :disabled="busy.accounts" @click="refreshAll">
        <RefreshCw :class="{ spin: busy.accounts || busy.aliases || busy.inbox }" :size="18" />
        <span>刷新</span>
      </button>
    </header>

    <section v-if="error || notice" class="feedback" :class="{ danger: error }">
      <AlertCircle v-if="error" :size="18" />
      <CheckCircle2 v-else :size="18" />
      <span>{{ error || notice }}</span>
    </section>

    <section class="workspace">
      <aside class="panel account-panel">
        <div class="panel-title">
          <ShieldCheck :size="20" />
          <span>账号</span>
        </div>

        <label class="field-label" for="account">当前账号</label>
        <select id="account" v-model="selectedAccountId" class="select" @change="handleAccountChange">
          <option v-for="account in accounts" :key="account.id" :value="account.id">
            {{ account.name || account.id }}
          </option>
        </select>

        <div v-if="activeAccount" class="account-card">
          <div class="status-line">
            <span class="status-dot" :class="{ active: activeAccount.status === 'active' }"></span>
            <strong>{{ activeAccount.status || "unknown" }}</strong>
          </div>
          <p>{{ activeAccount.icloud_email || activeAccount.real_email || "未设置邮箱" }}</p>
          <p>{{ activeAccount.host }}</p>
        </div>

        <div class="stats-grid">
          <div>
            <strong>{{ aliases.length }}</strong>
            <span>全部别名</span>
          </div>
          <div>
            <strong>{{ activeAliases }}</strong>
            <span>启用中</span>
          </div>
          <div>
            <strong>{{ inactiveAliases }}</strong>
            <span>已停用</span>
          </div>
        </div>
      </aside>

      <section class="panel alias-panel">
        <div class="panel-heading">
          <div class="panel-title">
            <Mail :size="20" />
            <span>隐私邮箱</span>
          </div>
          <button class="ghost-button" type="button" :disabled="busy.aliases" @click="loadAliases">
            <RefreshCw :class="{ spin: busy.aliases }" :size="16" />
            更新列表
          </button>
        </div>

        <form class="create-row" @submit.prevent="createAlias">
          <input
            v-model="newLabel"
            class="input"
            placeholder="新别名标签，例如 GitHub 注册"
            aria-label="新别名标签"
          />
          <button class="primary-button" type="submit" :disabled="busy.create || !selectedAccountId">
            <Loader2 v-if="busy.create" class="spin" :size="17" />
            <Plus v-else :size="17" />
            创建别名
          </button>
        </form>

        <div v-if="busy.aliases" class="empty-state">正在读取别名列表...</div>
        <div v-else-if="aliases.length === 0" class="empty-state">当前账号还没有隐私邮箱别名。</div>
        <div v-else class="alias-list">
          <button
            v-for="alias in aliases"
            :key="alias.anonymousId || alias.email"
            class="alias-row"
            :class="{ selected: selectedAlias === alias.email }"
            type="button"
            @click="chooseAlias(alias)"
          >
            <span class="alias-main">
              <strong>{{ alias.email }}</strong>
              <small>{{ alias.label || "未命名" }}</small>
            </span>
            <span class="alias-meta">
              <span class="tag" :class="{ muted: !alias.active }">
                {{ alias.active ? "启用" : "停用" }}
              </span>
              <small>{{ formatDate(alias.createdAt) }}</small>
            </span>
          </button>
        </div>
      </section>

      <section class="panel message-panel">
        <div class="panel-heading">
          <div class="panel-title">
            <Inbox :size="20" />
            <span>邮件</span>
          </div>
          <button class="ghost-button" type="button" :disabled="busy.inbox" @click="loadInbox()">
            <RefreshCw :class="{ spin: busy.inbox }" :size="16" />
            读取邮件
          </button>
        </div>

        <div class="mail-toolbar">
          <label class="folder-select">
            <Folder :size="16" />
            <select v-model="selectedFolder" class="select" :disabled="busy.folders" @change="loadInbox()">
              <option v-for="folder in folderOptions" :key="folder.name" :value="folder.name">
                {{ folderLabel(folder) }}
              </option>
            </select>
          </label>
        </div>

        <div class="mail-summary">
          <div>
            <strong>{{ inboxMeta.count }}</strong>
            <span>封邮件</span>
          </div>
          <div>
            <strong>{{ inboxMeta.method || "未读取" }}</strong>
            <span>{{ folderLabel(inboxMeta.folder) }}</span>
          </div>
        </div>

        <div v-if="selectedAliasInfo" class="selected-alias">
          <span>{{ selectedAliasInfo.email }}</span>
          <button class="mini-button" type="button" @click="copyText(selectedAliasInfo.email)">
            <Copy :size="14" />
            复制
          </button>
          <button class="mini-button" type="button" @click="clearAliasSelection">查看全部邮件</button>
        </div>

        <div v-if="busy.inbox" class="empty-state">正在读取邮件...</div>
        <div v-else-if="messages.length === 0" class="empty-state">
          {{ selectedAlias ? "这个别名在当前文件夹范围内未读取到邮件。" : "当前范围暂无可显示邮件。" }}
        </div>
        <article v-for="message in messages" v-else :key="message.id" class="message-card">
          <div class="message-top">
            <strong>{{ message.subject || "无主题" }}</strong>
            <time>{{ formatDate(message.date) }}</time>
          </div>
          <p class="message-from">{{ message.from }}</p>
          <p v-if="message.to" class="message-to">收件人：{{ message.to }}</p>
          <p v-if="message.folder" class="folder-badge">{{ folderLabel(message.folder) }}</p>
          <p class="message-preview">{{ message.preview || "无正文摘要" }}</p>
        </article>
      </section>
    </section>
  </main>
</template>
