const storageKeys = {
  accessToken: "realworld.accessToken",
  refreshToken: "realworld.refreshToken",
  currentUser: "realworld.currentUser",
};

const state = {
  accessToken: "",
  refreshToken: "",
  currentUser: null,
  readyStatus: "unknown",
  tags: [],
  articles: [],
  articlesTotal: 0,
  selectedSlug: "",
  currentArticle: null,
  currentComments: [],
  feedMode: false,
  editingSlug: "",
};

const els = {
  readyStatus: document.getElementById("ready-status"),
  sessionStatus: document.getElementById("session-status"),
  currentUserName: document.getElementById("current-user-name"),
  currentUserEmail: document.getElementById("current-user-email"),
  accessToken: document.getElementById("access-token"),
  refreshToken: document.getElementById("refresh-token"),
  registerForm: document.getElementById("register-form"),
  loginForm: document.getElementById("login-form"),
  articleForm: document.getElementById("article-form"),
  commentForm: document.getElementById("comment-form"),
  filtersForm: document.getElementById("filters-form"),
  articlesList: document.getElementById("articles-list"),
  articlesSummary: document.getElementById("articles-summary"),
  tagCloud: document.getElementById("tag-cloud"),
  emptyDetail: document.getElementById("empty-detail"),
  articleDetail: document.getElementById("article-detail"),
  detailSlug: document.getElementById("detail-slug"),
  detailTitle: document.getElementById("detail-title"),
  detailAuthor: document.getElementById("detail-author"),
  detailCreated: document.getElementById("detail-created"),
  detailUpdated: document.getElementById("detail-updated"),
  detailFavorites: document.getElementById("detail-favorites"),
  detailDescription: document.getElementById("detail-description"),
  detailBody: document.getElementById("detail-body"),
  detailTags: document.getElementById("detail-tags"),
  commentsSummary: document.getElementById("comments-summary"),
  commentsList: document.getElementById("comments-list"),
  refreshHealth: document.getElementById("refresh-health"),
  loadCurrentUser: document.getElementById("load-current-user"),
  refreshTokenButton: document.getElementById("refresh-token-button"),
  logoutButton: document.getElementById("logout-button"),
  reloadArticles: document.getElementById("reload-articles"),
  reloadTags: document.getElementById("reload-tags"),
  clearFilters: document.getElementById("clear-filters"),
  showGlobalFeed: document.getElementById("show-global-feed"),
  showPersonalFeed: document.getElementById("show-personal-feed"),
  refreshDetailButton: document.getElementById("refresh-detail-button"),
  favoriteButton: document.getElementById("favorite-button"),
  deleteArticleButton: document.getElementById("delete-article-button"),
  editorTitle: document.getElementById("editor-title"),
  articleSubmitButton: document.getElementById("article-submit-button"),
  cancelEditButton: document.getElementById("cancel-edit-button"),
  loadSelectedIntoEditor: document.getElementById("load-selected-into-editor"),
  loadArticleEditButton: document.getElementById("load-article-edit-button"),
  toast: document.getElementById("toast"),
};

let toastTimer;

document.addEventListener("DOMContentLoaded", boot);

async function boot() {
  restoreSession();
  bindEvents();
  syncSessionUI();
  syncEditorMode();
  applyHashSelection();

  await refreshHealth();
  await loadTags();
  await loadArticles();

  if (state.accessToken) {
    try {
      await loadCurrentUser(true);
    } catch (error) {
      console.warn(error);
    }
  }
}

function bindEvents() {
  els.registerForm.addEventListener("submit", handleRegister);
  els.loginForm.addEventListener("submit", handleLogin);
  els.articleForm.addEventListener("submit", handleArticleSubmit);
  els.commentForm.addEventListener("submit", handleCommentSubmit);
  els.filtersForm.addEventListener("submit", handleFilterSubmit);

  els.refreshHealth.addEventListener("click", () => refreshHealth(true));
  els.loadCurrentUser.addEventListener("click", () => loadCurrentUser());
  els.refreshTokenButton.addEventListener("click", handleRefreshToken);
  els.logoutButton.addEventListener("click", handleLogout);
  els.reloadArticles.addEventListener("click", () => loadArticles());
  els.reloadTags.addEventListener("click", () => loadTags(true));
  els.clearFilters.addEventListener("click", handleClearFilters);
  els.showGlobalFeed.addEventListener("click", () => {
    state.feedMode = false;
    loadArticles(true);
  });
  els.showPersonalFeed.addEventListener("click", () => {
    if (!state.accessToken) {
      showToast("先登录后才能查看个人 Feed。", "error");
      return;
    }
    state.feedMode = true;
    loadArticles(true);
  });
  els.refreshDetailButton.addEventListener("click", () => {
    if (state.selectedSlug) {
      loadArticle(state.selectedSlug, true);
    }
  });
  els.favoriteButton.addEventListener("click", toggleFavoriteCurrentArticle);
  els.deleteArticleButton.addEventListener("click", deleteCurrentArticle);
  els.cancelEditButton.addEventListener("click", cancelEditMode);
  els.loadSelectedIntoEditor.addEventListener("click", () => {
    if (!state.currentArticle) {
      showToast("先选择一篇文章再载入编辑器。", "error");
      return;
    }
    enterEditMode(state.currentArticle);
  });
  els.loadArticleEditButton.addEventListener("click", () => {
    if (!state.currentArticle) {
      return;
    }
    enterEditMode(state.currentArticle);
  });

  window.addEventListener("hashchange", applyHashSelection);
}

function restoreSession() {
  state.accessToken = localStorage.getItem(storageKeys.accessToken) || "";
  state.refreshToken = localStorage.getItem(storageKeys.refreshToken) || "";

  const currentUserRaw = localStorage.getItem(storageKeys.currentUser);
  if (currentUserRaw) {
    try {
      state.currentUser = JSON.parse(currentUserRaw);
    } catch (error) {
      console.warn("Failed to restore current user", error);
    }
  }
}

function persistSession(user, accessToken, refreshToken) {
  if (accessToken !== undefined) {
    state.accessToken = accessToken || "";
    localStorage.setItem(storageKeys.accessToken, state.accessToken);
  }

  if (refreshToken !== undefined) {
    state.refreshToken = refreshToken || "";
    localStorage.setItem(storageKeys.refreshToken, state.refreshToken);
  }

  if (user !== undefined) {
    state.currentUser = user;
    if (user) {
      localStorage.setItem(storageKeys.currentUser, JSON.stringify(user));
    } else {
      localStorage.removeItem(storageKeys.currentUser);
    }
  }

  syncSessionUI();
}

function clearSession() {
  localStorage.removeItem(storageKeys.accessToken);
  localStorage.removeItem(storageKeys.refreshToken);
  localStorage.removeItem(storageKeys.currentUser);
  state.accessToken = "";
  state.refreshToken = "";
  state.currentUser = null;
  syncSessionUI();
}

function syncSessionUI() {
  const user = state.currentUser;
  els.sessionStatus.textContent = user ? `已登录：${user.username}` : "未登录";
  els.currentUserName.textContent = user?.username || "未登录";
  els.currentUserEmail.textContent = user?.email || "-";
  els.accessToken.value = state.accessToken || "";
  els.refreshToken.value = state.refreshToken || "";
  els.refreshTokenButton.disabled = !state.refreshToken;
}

function syncEditorMode() {
  const editing = Boolean(state.editingSlug);
  els.editorTitle.textContent = editing ? `编辑文章：${state.editingSlug}` : "发布文章";
  els.articleSubmitButton.textContent = editing ? "保存修改" : "发布文章";
  els.cancelEditButton.hidden = !editing;
}

function currentUsername() {
  return state.currentUser?.username || "";
}

async function handleRegister(event) {
  event.preventDefault();
  const formData = new FormData(event.currentTarget);

  try {
    const response = await api("/api/users", {
      method: "POST",
      body: {
        user: {
          username: String(formData.get("username") || "").trim(),
          email: String(formData.get("email") || "").trim(),
          password: String(formData.get("password") || ""),
        },
      },
      auth: false,
    });

    applyAuthResponse(response.user);
    event.currentTarget.reset();
    showToast("注册成功，已经自动登录。", "success");
    await loadArticles();
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function handleLogin(event) {
  event.preventDefault();
  const formData = new FormData(event.currentTarget);

  try {
    const response = await api("/api/users/login", {
      method: "POST",
      body: {
        user: {
          email: String(formData.get("email") || "").trim(),
          password: String(formData.get("password") || ""),
        },
      },
      auth: false,
    });

    applyAuthResponse(response.user);
    event.currentTarget.reset();
    showToast("登录成功。", "success");
    await loadArticles();
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function loadCurrentUser(quiet = false) {
  try {
    const response = await api("/api/user");
    applyAuthResponse(response.user);
    if (!quiet) {
      showToast("当前用户信息已刷新。", "success");
    }
    return response.user;
  } catch (error) {
    if (String(error.message).includes("Unauthorized")) {
      clearSession();
    }
    if (!quiet) {
      showToast(error.message, "error");
    }
    throw error;
  }
}

async function handleRefreshToken() {
  if (!state.refreshToken) {
    showToast("当前没有 refresh token。", "error");
    return;
  }

  try {
    const response = await api("/api/users/refresh", {
      method: "POST",
      body: { refresh_token: state.refreshToken },
      auth: false,
    });

    persistSession(undefined, response.token, response.refresh_token);
    await loadCurrentUser(true);
    showToast("令牌已刷新。", "success");
  } catch (error) {
    showToast(error.message, "error");
  }
}

function handleLogout() {
  clearSession();
  showToast("已退出登录。", "success");
  loadArticles(true);
}

async function handleArticleSubmit(event) {
  event.preventDefault();

  if (!state.accessToken) {
    showToast("请先登录后再发文。", "error");
    return;
  }

  const formData = new FormData(event.currentTarget);
  const payload = {
    article: {
      title: String(formData.get("title") || "").trim(),
      description: String(formData.get("description") || "").trim(),
      body: String(formData.get("body") || "").trim(),
      tagList: parseTags(String(formData.get("tags") || "")),
    },
  };

  try {
    let response;
    if (state.editingSlug) {
      response = await api(`/api/articles/${encodeURIComponent(state.editingSlug)}`, {
        method: "PUT",
        body: payload,
      });
      showToast("文章已更新。", "success");
    } else {
      response = await api("/api/articles", {
        method: "POST",
        body: payload,
      });
      showToast("文章已发布。", "success");
    }

    cancelEditMode();
    event.currentTarget.reset();
    await loadTags(true);
    await loadArticles();
    await loadArticle(response.article.slug, true);
  } catch (error) {
    showToast(error.message, "error");
  }
}

function enterEditMode(article) {
  if (!article) {
    return;
  }

  state.editingSlug = article.slug;
  els.articleForm.elements.title.value = article.title || "";
  els.articleForm.elements.description.value = article.description || "";
  els.articleForm.elements.body.value = article.body || "";
  els.articleForm.elements.tags.value = (article.tagList || []).join(", ");
  syncEditorMode();
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function cancelEditMode() {
  state.editingSlug = "";
  els.articleForm.reset();
  syncEditorMode();
}

async function handleFilterSubmit(event) {
  event.preventDefault();
  state.feedMode = false;
  await loadArticles(true);
}

function handleClearFilters() {
  els.filtersForm.reset();
  els.filtersForm.elements.offset.value = 0;
  els.filtersForm.elements.limit.value = 20;
  state.feedMode = false;
  loadArticles(true);
}

async function loadArticles(resetOffset = false) {
  if (resetOffset) {
    els.filtersForm.elements.offset.value = 0;
  }

  els.articlesSummary.textContent = state.feedMode ? "正在加载我的 Feed…" : "正在加载文章…";

  try {
    const formData = new FormData(els.filtersForm);
    const params = new URLSearchParams();
    const limit = String(formData.get("limit") || "20").trim();
    const offset = String(formData.get("offset") || "0").trim();

    if (limit) {
      params.set("limit", limit);
    }
    if (offset) {
      params.set("offset", offset);
    }

    if (!state.feedMode) {
      const author = String(formData.get("author") || "").trim();
      const tag = String(formData.get("tag") || "").trim();
      const favorited = String(formData.get("favorited") || "").trim();

      if (author) {
        params.set("author", author);
      }
      if (tag) {
        params.set("tag", tag);
      }
      if (favorited) {
        params.set("favorited", favorited);
      }
    }

    const endpoint = state.feedMode ? "/api/articles/feed" : "/api/articles";
    const response = await api(`${endpoint}?${params.toString()}`);
    state.articles = response.articles || [];
    state.articlesTotal = response.articlesCount || 0;
    renderArticleList(state.articles, state.articlesTotal);

    if (state.selectedSlug && !state.articles.some((article) => article.slug === state.selectedSlug) && !state.currentArticle) {
      clearArticleDetail();
    }
  } catch (error) {
    state.articles = [];
    state.articlesTotal = 0;
    renderArticleList([], 0);
    showToast(error.message, "error");
  }
}

function renderArticleList(articles, total) {
  els.articlesList.innerHTML = "";
  els.articlesSummary.textContent = state.feedMode
    ? `我的 Feed 共 ${total} 篇`
    : `当前结果共 ${total} 篇`;

  if (!articles.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.innerHTML = "<p>当前没有文章，试试先登录后发布一篇，或者换个筛选条件。</p>";
    els.articlesList.appendChild(empty);
    return;
  }

  for (const article of articles) {
    const card = document.createElement("article");
    card.className = `article-card${article.slug === state.selectedSlug ? " active" : ""}`;
    card.dataset.slug = article.slug;
    card.addEventListener("click", () => loadArticle(article.slug, true));

    const title = document.createElement("h3");
    title.textContent = article.title;

    const description = document.createElement("p");
    description.textContent = article.description || "没有描述";

    const tags = document.createElement("div");
    tags.className = "chip-cloud";
    for (const tag of article.tagList || []) {
      const chip = createChip(tag, () => {
        els.filtersForm.elements.tag.value = tag;
        state.feedMode = false;
        loadArticles(true);
      });
      chip.addEventListener("click", (event) => event.stopPropagation());
      tags.appendChild(chip);
    }

    const meta = document.createElement("div");
    meta.className = "article-meta";
    meta.appendChild(createMetaPill(`作者 ${article.author?.username || "-"}`));
    meta.appendChild(createMetaPill(`收藏 ${article.favoritesCount || 0}`));
    meta.appendChild(createMetaPill(article.favorited ? "已收藏" : "未收藏"));
    meta.appendChild(createMetaPill(formatDate(article.updatedAt)));

    card.append(title, description, tags, meta);
    els.articlesList.appendChild(card);
  }
}

async function loadArticle(slug, updateHash = false) {
  if (!slug) {
    return;
  }

  state.selectedSlug = slug;
  if (updateHash) {
    window.location.hash = `article=${encodeURIComponent(slug)}`;
  }

  try {
    const [articleResponse, commentsResponse] = await Promise.all([
      api(`/api/articles/${encodeURIComponent(slug)}`),
      api(`/api/articles/${encodeURIComponent(slug)}/comments`),
    ]);

    state.currentArticle = articleResponse.article;
    state.currentComments = commentsResponse.comments || [];
    renderSelectedArticle();
    renderComments();
    highlightSelectedArticle();
  } catch (error) {
    showToast(error.message, "error");
  }
}

function renderSelectedArticle() {
  const article = state.currentArticle;
  if (!article) {
    clearArticleDetail();
    return;
  }

  els.emptyDetail.hidden = true;
  els.articleDetail.hidden = false;
  els.detailSlug.textContent = article.slug;
  els.detailTitle.textContent = article.title;
  els.detailAuthor.textContent = `作者：${article.author?.username || "-"}`;
  els.detailCreated.textContent = `创建：${formatDate(article.createdAt)}`;
  els.detailUpdated.textContent = `更新：${formatDate(article.updatedAt)}`;
  els.detailFavorites.textContent = `收藏：${article.favoritesCount || 0}`;
  els.detailDescription.textContent = article.description || "没有描述";
  els.detailBody.textContent = article.body || "";
  els.favoriteButton.textContent = article.favorited ? "取消收藏" : "收藏";
  els.favoriteButton.disabled = !state.accessToken;

  const ownArticle = currentUsername() && currentUsername() === article.author?.username;
  els.deleteArticleButton.hidden = !ownArticle;
  els.loadArticleEditButton.hidden = !ownArticle;

  els.detailTags.innerHTML = "";
  for (const tag of article.tagList || []) {
    els.detailTags.appendChild(createChip(tag, () => {
      els.filtersForm.elements.tag.value = tag;
      state.feedMode = false;
      loadArticles(true);
    }));
  }
}

function clearArticleDetail() {
  state.currentArticle = null;
  state.currentComments = [];
  state.selectedSlug = "";
  els.emptyDetail.hidden = false;
  els.articleDetail.hidden = true;
  els.commentsList.innerHTML = "";
  history.replaceState(null, "", window.location.pathname);
  highlightSelectedArticle();
}

function highlightSelectedArticle() {
  for (const card of els.articlesList.querySelectorAll(".article-card")) {
    card.classList.remove("active");
    if (card.dataset.slug === state.selectedSlug) {
      card.classList.add("active");
    }
  }
}

async function toggleFavoriteCurrentArticle() {
  if (!state.currentArticle) {
    return;
  }
  if (!state.accessToken) {
    showToast("登录后才能收藏文章。", "error");
    return;
  }

  const article = state.currentArticle;
  const method = article.favorited ? "DELETE" : "POST";

  try {
    const response = await api(`/api/articles/${encodeURIComponent(article.slug)}/favorite`, { method });
    state.currentArticle = response.article;
    syncArticleInList(response.article);
    renderSelectedArticle();
    renderArticleList(state.articles, state.articlesTotal);
    highlightSelectedArticle();
    showToast(article.favorited ? "已取消收藏。" : "已收藏文章。", "success");
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function deleteCurrentArticle() {
  if (!state.currentArticle) {
    return;
  }

  if (!confirm(`确定删除《${state.currentArticle.title}》吗？`)) {
    return;
  }

  try {
    await api(`/api/articles/${encodeURIComponent(state.currentArticle.slug)}`, { method: "DELETE" });
    showToast("文章已删除。", "success");
    cancelEditMode();
    clearArticleDetail();
    await loadTags(true);
    await loadArticles();
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function handleCommentSubmit(event) {
  event.preventDefault();

  if (!state.currentArticle) {
    showToast("先选择一篇文章。", "error");
    return;
  }
  if (!state.accessToken) {
    showToast("登录后才能评论。", "error");
    return;
  }

  const formData = new FormData(event.currentTarget);
  const body = String(formData.get("body") || "").trim();
  if (!body) {
    showToast("评论内容不能为空。", "error");
    return;
  }

  try {
    await api(`/api/articles/${encodeURIComponent(state.currentArticle.slug)}/comments`, {
      method: "POST",
      body: { comment: { body } },
    });

    event.currentTarget.reset();
    await loadArticle(state.currentArticle.slug, false);
    showToast("评论已提交。", "success");
  } catch (error) {
    showToast(error.message, "error");
  }
}

function renderComments() {
  els.commentsList.innerHTML = "";
  els.commentsSummary.textContent = `${state.currentComments.length} 条评论`;

  if (!state.currentComments.length) {
    const empty = document.createElement("div");
    empty.className = "empty-state";
    empty.innerHTML = "<p>这篇文章暂时还没有评论。</p>";
    els.commentsList.appendChild(empty);
    return;
  }

  for (const comment of state.currentComments) {
    const card = document.createElement("article");
    card.className = "comment-card";

    const meta = document.createElement("div");
    meta.className = "comment-meta";
    meta.appendChild(createMetaPill(`作者 ${comment.author?.username || "-"}`));
    meta.appendChild(createMetaPill(formatDate(comment.createdAt)));

    const body = document.createElement("p");
    body.textContent = comment.body;

    card.append(meta, body);

    if (currentUsername() && currentUsername() === comment.author?.username) {
      const actions = document.createElement("div");
      actions.className = "button-row";
      const deleteButton = document.createElement("button");
      deleteButton.type = "button";
      deleteButton.className = "ghost-button";
      deleteButton.textContent = "删除评论";
      deleteButton.addEventListener("click", () => deleteComment(comment.id));
      actions.appendChild(deleteButton);
      card.appendChild(actions);
    }

    els.commentsList.appendChild(card);
  }
}

async function deleteComment(commentID) {
  if (!state.currentArticle) {
    return;
  }

  try {
    await api(`/api/articles/${encodeURIComponent(state.currentArticle.slug)}/comments/${commentID}`, {
      method: "DELETE",
    });
    await loadArticle(state.currentArticle.slug, false);
    showToast("评论已删除。", "success");
  } catch (error) {
    showToast(error.message, "error");
  }
}

async function loadTags(quiet = false) {
  try {
    const response = await api("/api/tags", { auth: false });
    state.tags = response.tags || [];
    renderTags();
    if (quiet) {
      showToast("标签已刷新。", "success");
    }
  } catch (error) {
    if (!quiet) {
      showToast(error.message, "error");
    }
  }
}

function renderTags() {
  els.tagCloud.innerHTML = "";

  if (!state.tags.length) {
    const chip = document.createElement("span");
    chip.className = "chip passive";
    chip.textContent = "暂无标签";
    els.tagCloud.appendChild(chip);
    return;
  }

  for (const tag of state.tags) {
    els.tagCloud.appendChild(createChip(tag, () => {
      els.filtersForm.elements.tag.value = tag;
      state.feedMode = false;
      loadArticles(true);
    }));
  }
}

async function refreshHealth(quiet = false) {
  try {
    const response = await api("/readyz", { auth: false });
    state.readyStatus = response.status || "ready";
    els.readyStatus.textContent = response.status === "ready" ? "后端就绪" : JSON.stringify(response);
    if (quiet) {
      showToast("健康状态已刷新。", "success");
    }
  } catch (error) {
    state.readyStatus = "failed";
    els.readyStatus.textContent = "不可用";
    if (!quiet) {
      showToast(error.message, "error");
    }
  }
}

function applyAuthResponse(userPayload) {
  const user = {
    username: userPayload.username,
    email: userPayload.email,
    bio: userPayload.bio || null,
    image: userPayload.image || null,
  };

  persistSession(user, userPayload.token || state.accessToken, userPayload.refresh_token || state.refreshToken);
}

function syncArticleInList(article) {
  const index = state.articles.findIndex((item) => item.slug === article.slug);
  if (index >= 0) {
    state.articles[index] = article;
  }
}

function parseTags(value) {
  return value
    .split(/[\n,]/)
    .map((tag) => tag.trim())
    .filter(Boolean);
}

function createChip(label, onClick) {
  const chip = document.createElement("button");
  chip.type = "button";
  chip.className = "chip";
  chip.textContent = label;
  if (onClick) {
    chip.addEventListener("click", onClick);
  }
  return chip;
}

function createMetaPill(text) {
  const pill = document.createElement("span");
  pill.className = "chip passive";
  pill.textContent = text;
  return pill;
}

function formatDate(value) {
  if (!value) {
    return "-";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function applyHashSelection() {
  const hash = window.location.hash.replace(/^#/, "");
  if (!hash.startsWith("article=")) {
    return;
  }

  const slug = decodeURIComponent(hash.slice("article=".length));
  if (slug && slug !== state.selectedSlug) {
    loadArticle(slug, false);
  }
}

async function api(path, options = {}) {
  const {
    method = "GET",
    body,
    headers = {},
    auth = true,
  } = options;

  const requestHeaders = {
    Accept: "application/json",
    ...headers,
  };

  if (auth && state.accessToken) {
    requestHeaders.Authorization = `Token ${state.accessToken}`;
  }

  const response = await fetch(path, {
    method,
    headers: body === undefined ? requestHeaders : {
      "Content-Type": "application/json",
      ...requestHeaders,
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });

  const raw = await response.text();
  const payload = raw ? tryParseJSON(raw) : null;
  const refreshToken = response.headers.get("X-Refresh-Token");

  if (refreshToken && payload && typeof payload === "object" && payload.user && typeof payload.user === "object") {
    payload.user.refresh_token = refreshToken;
  }

  if (!response.ok) {
    throw new Error(extractErrorMessage(payload, response.status));
  }

  return payload;
}

function tryParseJSON(raw) {
  try {
    return JSON.parse(raw);
  } catch (error) {
    return raw;
  }
}

function extractErrorMessage(payload, status) {
  if (payload && typeof payload === "object") {
    if (typeof payload.error === "string") {
      return payload.error;
    }
    if (payload.errors && typeof payload.errors === "object") {
      return Object.values(payload.errors).join("；");
    }
  }

  if (typeof payload === "string" && payload.trim()) {
    return payload;
  }

  return `请求失败（HTTP ${status}）`;
}

function showToast(message, type = "success") {
  clearTimeout(toastTimer);
  els.toast.hidden = false;
  els.toast.className = `toast ${type}`;
  els.toast.textContent = message;

  toastTimer = window.setTimeout(() => {
    els.toast.hidden = true;
  }, 2600);
}
