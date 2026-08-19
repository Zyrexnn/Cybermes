/* =============================================================================
   app.js - HackSkills main page logic
   - 加载 data/skills.json
   - 渲染分类侧栏 + skill 卡片 (按分类分节 / 搜索时按相关度扁平)
   - 协调 hero 主搜索 + 顶部 sticky 辅搜索 (双向同步)
   - install tab 切换 + per-skill 命令更新 + per-card copy
   - URL hash 路由 + 键盘快捷键 / Esc 1-9
   关键词: data loading, render pipeline, hero/toolbar dual search, install command updater
   ============================================================================= */
(function () {
  "use strict";

  const DATA_URL = "data/skills.json";

  const els = {
    cards:           document.getElementById("cards"),
    cardTpl:         document.getElementById("card-template"),
    catNav:          document.getElementById("category-nav"),
    loading:         document.getElementById("loading"),
    empty:           document.getElementById("empty"),
    error:           document.getElementById("error"),
    toolbar:         document.getElementById("toolbar"),
    chips:           document.querySelectorAll(".chip"),
    counter:         document.getElementById("result-counter"),
    heroSearchForm:  document.getElementById("hero-search-form"),
    heroSearchInput: document.getElementById("hero-search-input"),
    searchInput:     document.getElementById("search-input"),
    searchClear:     document.getElementById("search-clear"),
    toolbarSearch:   document.querySelector(".toolbar-search"),
    heroVersion:     document.getElementById("hero-version"),
    heroSkillsNum:   document.getElementById("hero-skills-num"),
    heroCatsNum:     document.getElementById("hero-cats-num"),
    heroLines:       document.getElementById("hero-lines"),
    heroBuild:       document.getElementById("hero-build"),
    brandVersion:    document.getElementById("brand-version"),
    provSkillsNum:   document.getElementById("prov-skills-num"),
    provCatsNum:     document.getElementById("prov-cats-num"),
    provBuild:       document.getElementById("prov-build"),
    marqueeTrack:    document.getElementById("marquee-track"),
    footVersion:     document.getElementById("foot-version"),
    footBuildTime:   document.getElementById("foot-build-time"),
    footYear:        document.getElementById("foot-year"),
    installTabs:     document.querySelectorAll(".install-tab"),
    installPanels:   document.querySelectorAll(".install-panel"),
    singleSelect:    document.getElementById("single-skill-select"),
    curlSelect:      document.getElementById("curl-skill-select"),
    agentSelect:     document.getElementById("agent-skill-select"),
    cmdSingleArg:    document.getElementById("cmd-single-arg"),
    cmdCurlArg:      document.getElementById("cmd-curl-arg"),
    agentName:       document.getElementById("agent-name"),
    agentUrl:        document.getElementById("agent-url"),
    agentBlob:       document.getElementById("agent-blob"),
    navDownload:     document.getElementById("nav-download"),
    zipModal:        document.getElementById("zip-modal"),
    zipModalClose:   document.getElementById("zip-modal-close"),
    modalZipVersion: document.getElementById("modal-zip-version"),
    modalZipDlPrimary: document.getElementById("modal-zip-dl-primary")
  };

  // 状态
  // 关键词: app state
  const state = {
    raw: null,
    skills: [],
    categories: [],
    activeCategory: null, // null = all
    activeTier: "all",
    query: "",
    githubRepo: "yaklang/hack-skills",
    githubBranch: "main"
  };

  // ---------------------------------------------------------------
  // utility: friendly title formatter (kebab -> Title Case)
  // ---------------------------------------------------------------
  function prettyTitle(id) {
    if (!id) return "";
    return id
      .split("-")
      .map(p => {
        if (/^[a-z]+\d/.test(p) || /^\d/.test(p)) return p.toUpperCase();
        if (p.length <= 3) return p.toUpperCase();
        return p.charAt(0).toUpperCase() + p.slice(1);
      })
      .join(" ");
  }

  function tierLabel(t) {
    return t === "master" ? "P0 master"
         : t === "category" ? "P1 category"
         :                    "P2 topic";
  }

  function formatNum(n) {
    return (n | 0).toLocaleString();
  }

  // ---------------------------------------------------------------
  // 初始加载
  // ---------------------------------------------------------------
  fetch(DATA_URL, { cache: "no-cache" })
    .then(r => { if (!r.ok) throw new Error("HTTP " + r.status); return r.json(); })
    .then(json => {
      state.raw = json;
      state.categories = json.categories || [];
      state.githubRepo = json.githubRepo || state.githubRepo;
      state.githubBranch = json.githubBranch || state.githubBranch;
      state.skills = [];
      for (const cat of state.categories) {
        for (const s of (cat.skills || [])) state.skills.push(s);
      }
      els.loading.classList.add("hidden");
      bootstrap();
    })
    .catch(err => {
      console.error("[hackskills] failed to load skills.json:", err);
      els.loading.classList.add("hidden");
      els.error.classList.remove("hidden");
    });

  function bootstrap() {
    renderHero();
    renderMarquee();
    renderCategoryNav();
    populateInstallSelects();
    bindHeroSearch();
    bindToolbarSearch();
    bindChips();
    bindKeyboard();
    bindHashRouter();
    bindInstallTabs();
    bindInstallSelects();
    bindZipModal();
    applyHashState();
    rerender();
  }

  // ---------------------------------------------------------------
  // hero / footer 静态信息
  // ---------------------------------------------------------------
  function renderHero() {
    const r = state.raw;
    els.heroVersion.textContent   = r.version || "dev";
    els.heroSkillsNum.textContent = formatNum(r.totalSkills);
    els.heroCatsNum.textContent   = formatNum(r.totalCategories);
    els.heroLines.textContent     = formatNum(r.totalLines);
    els.heroBuild.textContent     = (r.buildTime || "").replace("T", " ").replace("Z", " UTC");
    els.footVersion.textContent   = r.version || "—";
    els.footBuildTime.textContent = (r.buildTime || "").replace("T", " ").replace("Z", " UTC");
    if (els.footYear) els.footYear.textContent = String(new Date().getFullYear());

    // brand-version 徽章 (与 hackbenchmark 同款), provenance bar
    // 关键词: brand-version, provenance bar, sister site visual parity
    if (els.brandVersion) {
      const v = r.version || "dev";
      els.brandVersion.textContent = "v" + v.split("-").slice(0, 2).join("-");
    }
    if (els.provSkillsNum) els.provSkillsNum.textContent = formatNum(r.totalSkills);
    if (els.provCatsNum)   els.provCatsNum.textContent   = formatNum(r.totalCategories);
    if (els.provBuild)     els.provBuild.textContent     = r.version || "dev";

    // hero 占位 placeholder 数字也跟着更新
    if (els.heroSearchInput) {
      els.heroSearchInput.placeholder = `Search ${formatNum(r.totalSkills)} skills…   try "xss waf"   "category:auth"   "tier:deep"`;
    }
  }

  // ---------------------------------------------------------------
  // Hero skills marquee — 精选 skill chip 跑马灯, 点击直接过滤
  // 关键词: featured skills picking, marquee population, click-to-filter
  // ---------------------------------------------------------------
  function pickFeaturedSkills() {
    // 选取策略:
    //   1. 全部 master + category tier (路由型, 优先曝光)
    //   2. 每个 category 挑 1 个最长的 deep skill (代表性 deep)
    // 关键词: featured selection, master/category/top-deep
    const out = [];
    const seen = new Set();
    const push = (s) => {
      if (!s || seen.has(s.id)) return;
      seen.add(s.id);
      out.push(s);
    };

    const master = state.skills.find(s => s.tier === "master");
    if (master) push(master);

    for (const cat of state.categories) {
      const catTierSkill = cat.skills.find(s => s.tier === "category");
      if (catTierSkill) push(catTierSkill);
    }
    for (const cat of state.categories) {
      const deeps = cat.skills.filter(s => s.tier === "deep");
      if (deeps.length === 0) continue;
      const top = deeps.reduce((a, b) => (a.skillLines > b.skillLines ? a : b));
      push(top);
    }
    return out;
  }

  // 提取 description 的核心摘要片段, 去掉模板化的 "XXX playbook. Use when " 前缀
  // 关键词: chip summary extraction, description trimming
  function chipSummary(desc, max) {
    if (!desc) return "";
    const limit = typeof max === "number" ? max : 95;
    let s = String(desc).trim();
    const original = s;
    // 跳过模板化的第一句 (e.g. "XXX playbook." / "Entry P0 router for ...")
    // 关键词: skip preamble sentence, keep substantive scenario
    const dotIdx = s.indexOf(". ");
    if (dotIdx > 0 && dotIdx < s.length - 4) {
      const tail = s.slice(dotIdx + 2).trim();
      if (tail.length >= 12) s = tail;
    }
    s = s.replace(/^Use\s+when\s+/i, "");
    s = s.replace(/^[A-Z]/, m => m.toLowerCase());
    if (!s || s.length < 12) s = original;
    s = s.replace(/[\s.]+$/, "");
    if (s.length > limit) {
      s = s.slice(0, limit).replace(/\s+\S*$/, "") + "…";
    }
    return s;
  }

  // skill id 关键词标签: 取末尾 1-2 个有信息量的 token
  // 关键词: chip tags from skill id, stopword filter
  const CHIP_TAG_STOPWORDS = new Set([
    "and","for","the","with","via","from","into","out","vs",
    "app","apps","based","using","deep","attacks","attack"
  ]);
  function chipTags(id) {
    if (!id) return [];
    const parts = id.split("-").filter(p => p.length >= 3 && !CHIP_TAG_STOPWORDS.has(p));
    if (parts.length === 0) return [];
    return parts.slice(-2);
  }

  function buildChipNode(skill) {
    const a = document.createElement("a");
    a.href = "#browse";
    a.className = "marquee-chip";
    a.dataset.tier = skill.tier;
    a.dataset.skillId = skill.id;
    a.title = `${skill.id}\n${skill.description || ""}`;
    const tierMark = skill.tier === "master" ? "P0 master"
                  : skill.tier === "category" ? "P1 category"
                  : "P2 topic";
    const cat = state.categories.find(c => c.id === skill.category);
    const catName = cat ? cat.name : skill.category;
    const summary = chipSummary(skill.description, 95);
    const tags = chipTags(skill.id);
    const tagsHtml = tags.length
      ? tags.map(t => `<span class="chip-tag">${HackSkillsSearch.escapeHtml(t)}</span>`).join("")
      : "";
    const docsCount = (skill.extraDocs && skill.extraDocs.length) || 0;
    const metric = docsCount > 0
      ? `${formatNum(skill.skillLines)}L · ${docsCount} doc${docsCount > 1 ? "s" : ""}`
      : `${formatNum(skill.skillLines)}L`;
    a.innerHTML = `
      <div class="chip-row-1">
        <span class="chip-mark">${tierMark}</span>
        <span class="chip-id">${HackSkillsSearch.escapeHtml(skill.id)}</span>
        <span class="chip-metric">${metric}</span>
      </div>
      <div class="chip-row-2">
        <span class="chip-cat">${HackSkillsSearch.escapeHtml(catName)}</span>
        <span class="chip-sep">·</span>
        <span class="chip-summary">${HackSkillsSearch.escapeHtml(summary)}</span>
        ${tagsHtml}
      </div>`;
    a.addEventListener("click", (e) => {
      e.preventDefault();
      // 把 skill id 作为搜索词填入并滚到 browse, 让用户直观看到结果
      // 关键词: chip click route, hero-to-browse jump
      const q = skill.id;
      els.heroSearchInput.value = q;
      els.searchInput.value = q;
      els.toolbarSearch.classList.add("has-text");
      state.query = q;
      state.activeCategory = null;
      state.activeTier = "all";
      els.chips.forEach(c => c.setAttribute("aria-pressed", c.dataset.tier === "all" ? "true" : "false"));
      syncCategoryNav();
      syncHash();
      rerender();
      const target = document.getElementById("browse");
      if (target) target.scrollIntoView({ behavior: "smooth", block: "start" });
    });
    return a;
  }

  function renderMarquee() {
    if (!els.marqueeTrack) return;
    const featured = pickFeaturedSkills();
    if (featured.length === 0) return;

    els.marqueeTrack.innerHTML = "";
    // 双倍内容拼接以实现无缝循环 (CSS 把轨道平移 -50%)
    // 关键词: seamless infinite scroll, double content
    for (let pass = 0; pass < 2; pass++) {
      for (const s of featured) {
        els.marqueeTrack.appendChild(buildChipNode(s));
      }
    }
  }

  // ---------------------------------------------------------------
  // 分类侧栏
  // ---------------------------------------------------------------
  function renderCategoryNav() {
    els.catNav.innerHTML = "";

    const allBtn = document.createElement("button");
    allBtn.type = "button";
    allBtn.className = "cat-item";
    allBtn.dataset.catId = "__all";
    allBtn.innerHTML = `
      <span class="cat-name">
        <span class="cat-num">·</span>
        <span class="cat-label">All skills</span>
      </span>
      <span class="cat-count">${state.skills.length}</span>`;
    allBtn.addEventListener("click", () => selectCategory(null));
    els.catNav.appendChild(allBtn);

    state.categories.forEach((cat, idx) => {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "cat-item";
      btn.dataset.catId = cat.id;
      btn.title = cat.description || cat.name;
      btn.innerHTML = `
        <span class="cat-name">
          <span class="cat-num">${String(idx + 1).padStart(2, "0")}</span>
          <span class="cat-label">${HackSkillsSearch.escapeHtml(cat.name)}</span>
        </span>
        <span class="cat-count">${cat.skillCount | 0}</span>`;
      btn.addEventListener("click", () => selectCategory(cat.id));
      els.catNav.appendChild(btn);
    });
    syncCategoryNav();
  }

  function syncCategoryNav() {
    const items = els.catNav.querySelectorAll(".cat-item");
    items.forEach(b => {
      const cid = b.dataset.catId;
      const isActive = (state.activeCategory === null && cid === "__all") ||
                       (state.activeCategory === cid);
      b.setAttribute("aria-selected", isActive ? "true" : "false");
    });
  }

  function selectCategory(cid) {
    state.activeCategory = cid;
    syncCategoryNav();
    syncHash();
    rerender();
    const target = document.getElementById("browse");
    if (target) target.scrollIntoView({ behavior: "smooth", block: "start" });
  }

  function selectTier(t) {
    state.activeTier = t;
    els.chips.forEach(c => c.setAttribute("aria-pressed", c.dataset.tier === t ? "true" : "false"));
    syncHash();
    rerender();
  }

  // ---------------------------------------------------------------
  // 双搜索框 (hero + toolbar) 双向同步
  // 关键词: dual search input sync
  // ---------------------------------------------------------------
  let debounceTimer = null;
  function applyQuery(v, src) {
    state.query = v;
    if (src !== "hero") els.heroSearchInput.value = v;
    if (src !== "toolbar") els.searchInput.value = v;
    if (v) els.toolbarSearch.classList.add("has-text");
    else   els.toolbarSearch.classList.remove("has-text");
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      syncHash();
      rerender();
    }, 80);
  }

  function bindHeroSearch() {
    els.heroSearchInput.addEventListener("input", (e) => applyQuery(e.target.value, "hero"));
    els.heroSearchForm.addEventListener("submit", (e) => {
      e.preventDefault();
      const target = document.getElementById("browse");
      if (target) target.scrollIntoView({ behavior: "smooth", block: "start" });
    });
  }

  function bindToolbarSearch() {
    els.searchInput.addEventListener("input", (e) => applyQuery(e.target.value, "toolbar"));
    els.searchClear.addEventListener("click", () => {
      applyQuery("", "clear");
      els.searchInput.focus();
    });
  }

  function bindChips() {
    els.chips.forEach(chip => {
      chip.addEventListener("click", () => selectTier(chip.dataset.tier));
    });
  }

  // ---------------------------------------------------------------
  // 键盘快捷键
  // ---------------------------------------------------------------
  function bindKeyboard() {
    document.addEventListener("keydown", (e) => {
      const inField = e.target && (e.target.tagName === "INPUT" || e.target.tagName === "TEXTAREA" || e.target.tagName === "SELECT");

      if (e.key === "/" && !inField) {
        e.preventDefault();
        const visible = isToolbarVisible() ? els.searchInput : els.heroSearchInput;
        visible.focus();
        visible.select();
        return;
      }
      if (e.key === "Escape" && inField) {
        if (e.target === els.heroSearchInput || e.target === els.searchInput) {
          applyQuery("", "esc");
          e.target.blur();
        }
        return;
      }
      if (inField) return;

      if (/^[1-9]$/.test(e.key)) {
        const idx = parseInt(e.key, 10) - 1;
        const cat = state.categories[idx];
        if (cat) selectCategory(cat.id);
        return;
      }
      if (e.key === "0") selectCategory(null);
    });
  }

  function isToolbarVisible() {
    const r = els.toolbar.getBoundingClientRect();
    return r.top < window.innerHeight && r.bottom > 0;
  }

  // ---------------------------------------------------------------
  // hash 路由
  // ---------------------------------------------------------------
  function bindHashRouter() {
    window.addEventListener("hashchange", applyHashState);
  }

  function applyHashState() {
    const hash = window.location.hash.replace(/^#/, "");
    if (!hash) return;
    if (hash === "install" || hash === "browse" || hash === "download") return; // anchor / modal trigger
    const params = new URLSearchParams(hash);
    const c = params.get("category");
    const t = params.get("tier");
    const q = params.get("q");
    if (c) state.activeCategory = (c === "all" ? null : c);
    if (t) {
      state.activeTier = t;
      els.chips.forEach(chip => chip.setAttribute("aria-pressed", chip.dataset.tier === t ? "true" : "false"));
    }
    if (q !== null) {
      state.query = q;
      els.searchInput.value = q;
      els.heroSearchInput.value = q;
      if (q) els.toolbarSearch.classList.add("has-text");
    }
    syncCategoryNav();
    rerender();
  }

  function syncHash() {
    const cur = window.location.hash;
    if (cur === "#install" || cur === "#browse" || cur === "#download") return;
    const params = new URLSearchParams();
    if (state.activeCategory !== null) params.set("category", state.activeCategory);
    if (state.activeTier !== "all")    params.set("tier", state.activeTier);
    if (state.query)                   params.set("q", state.query);
    const newHash = params.toString();
    const target = newHash ? "#" + newHash : "";
    if (window.location.hash !== target) {
      history.replaceState(null, "", target || (window.location.pathname + window.location.search));
    }
  }

  // ---------------------------------------------------------------
  // 主渲染
  // ---------------------------------------------------------------
  function rerender() {
    let virtual = state.query.trim();
    if (state.activeCategory) {
      if (!/(^|\s)(category|cat):/i.test(virtual)) {
        virtual = (virtual + " category:" + state.activeCategory).trim();
      }
    }
    if (state.activeTier !== "all") {
      if (!/(^|\s)tier:/i.test(virtual)) {
        virtual = (virtual + " tier:" + state.activeTier).trim();
      }
    }

    const r = HackSkillsSearch.search(state.skills, virtual);

    const count = r.results.length;
    els.counter.innerHTML = count === 0
      ? `<strong>0</strong> matches`
      : `<strong>${count}</strong> match${count === 1 ? "" : "es"}`;

    if (count === 0) {
      els.empty.classList.remove("hidden");
      els.cards.innerHTML = "";
      return;
    }
    els.empty.classList.add("hidden");

    const tokens = r.query.free;
    if (tokens.length === 0) {
      renderByCategories(r.results);
    } else {
      renderFlat(r.results, tokens);
    }
  }

  function renderByCategories(results) {
    els.cards.innerHTML = "";
    const grouped = new Map();
    for (const item of results) grouped.set(item.skill.category, []);
    for (const item of results) grouped.get(item.skill.category).push(item.skill);

    state.categories.forEach((cat, idx) => {
      const arr = grouped.get(cat.id);
      if (!arr || arr.length === 0) return;

      const header = document.createElement("div");
      header.className = "cat-section-header";
      header.innerHTML = `
        <span class="h-num">${String(idx + 1).padStart(2, "0")}</span>
        <span class="h-name">${HackSkillsSearch.escapeHtml(cat.name)}</span>
        <span class="h-count">${arr.length}</span>
        <span class="h-desc">${HackSkillsSearch.escapeHtml(cat.description || "")}</span>`;
      els.cards.appendChild(header);
      for (const s of arr) els.cards.appendChild(renderCard(s, []));
    });
  }

  function renderFlat(results, tokens) {
    els.cards.innerHTML = "";
    for (const item of results) els.cards.appendChild(renderCard(item.skill, tokens));
  }

  // ---------------------------------------------------------------
  // 单卡渲染
  // ---------------------------------------------------------------
  function buildInstallSnippet(skill) {
    return `npx skills add ${state.githubRepo}/${skill.id}`;
  }

  function renderCard(skill, highlightTokens) {
    const node = els.cardTpl.content.firstElementChild.cloneNode(true);
    const tier = skill.tier || "deep";

    const tierEl = node.querySelector(".card-tier");
    tierEl.dataset.tier = tier;
    tierEl.textContent = tierLabel(tier);

    const catLabel = node.querySelector(".card-cat-label");
    const catObj = state.categories.find(c => c.id === skill.category);
    catLabel.textContent = catObj ? catObj.name : (skill.category || "uncategorized");

    const title = node.querySelector(".card-name");
    title.innerHTML = HackSkillsSearch.highlight(prettyTitle(skill.id), highlightTokens);

    const titleLink = node.querySelector(".card-title-link");
    titleLink.href = skill.blobUrl;

    const cardId = node.querySelector(".card-id");
    cardId.innerHTML = HackSkillsSearch.highlight(skill.id, highlightTokens);

    const desc = node.querySelector(".card-desc");
    desc.innerHTML = HackSkillsSearch.highlight(skill.description || "", highlightTokens);

    const lines = node.querySelector(".card-lines");
    lines.textContent = formatNum(skill.skillLines);

    const extraStat = node.querySelector(".card-extra");
    const extraCount = node.querySelector(".card-extra-count");
    if (Array.isArray(skill.extraDocs) && skill.extraDocs.length > 0) {
      extraStat.classList.remove("hidden");
      extraCount.textContent = String(skill.extraDocs.length);
    }

    const open = node.querySelector(".card-open");
    open.href = skill.blobUrl;
    open.title = "Open " + skill.path + " on GitHub";

    const copyBtn = node.querySelector(".card-copy");
    const snippet = buildInstallSnippet(skill);
    copyBtn.dataset.copyText = snippet;
    copyBtn.title = "Copy: " + snippet;

    // 整卡点击 (避开内部链接 / 按钮)
    node.addEventListener("click", (e) => {
      if (e.target.closest("a, button")) return;
      window.open(skill.blobUrl, "_blank", "noopener,noreferrer");
    });
    node.addEventListener("keydown", (e) => {
      if ((e.key === "Enter" || e.key === " ") && !e.target.closest("button, a")) {
        e.preventDefault();
        window.open(skill.blobUrl, "_blank", "noopener,noreferrer");
      }
    });

    return node;
  }

  // ---------------------------------------------------------------
  // Install tab + per-skill 命令更新
  // 关键词: install tab switching, dynamic command rendering
  // ---------------------------------------------------------------
  function bindInstallTabs() {
    els.installTabs.forEach(tab => {
      tab.addEventListener("click", () => {
        const target = tab.dataset.tab;
        els.installTabs.forEach(t => t.setAttribute("aria-selected", t === tab ? "true" : "false"));
        els.installPanels.forEach(p => {
          if (p.dataset.panel === target) p.removeAttribute("hidden");
          else p.setAttribute("hidden", "");
        });
      });
    });
  }

  function populateInstallSelects() {
    const selects = [els.singleSelect, els.curlSelect, els.agentSelect].filter(Boolean);
    if (selects.length === 0) return;

    // 优先按 tier 排列: master 第一, 其次 category, 然后 deep (按分类顺序)
    const ordered = [];
    const masters   = state.skills.filter(s => s.tier === "master");
    const cats      = state.skills.filter(s => s.tier === "category");
    const deeps     = state.skills.filter(s => s.tier === "deep");
    ordered.push(...masters, ...cats, ...deeps);

    const html = ordered.map(s => {
      const tag = s.tier === "master" ? "[master] "
              : s.tier === "category" ? "[category] "
              :                         "";
      return `<option value="${HackSkillsSearch.escapeHtml(s.id)}">${tag}${HackSkillsSearch.escapeHtml(s.id)}</option>`;
    }).join("");

    selects.forEach(sel => {
      sel.innerHTML = html;
      sel.value = "hack";
    });
  }

  function bindInstallSelects() {
    if (els.singleSelect) {
      els.singleSelect.addEventListener("change", (e) => updateSingleCmd(e.target.value));
      updateSingleCmd(els.singleSelect.value);
    }
    if (els.curlSelect) {
      els.curlSelect.addEventListener("change", (e) => updateCurlCmd(e.target.value));
      updateCurlCmd(els.curlSelect.value);
    }
    if (els.agentSelect) {
      els.agentSelect.addEventListener("change", (e) => updateAgent(e.target.value));
      updateAgent(els.agentSelect.value);
    }
  }

  function findSkill(id) {
    return state.skills.find(s => s.id === id);
  }

  function updateSingleCmd(id) {
    if (!els.cmdSingleArg) return;
    els.cmdSingleArg.textContent = id || "hack";
  }

  function updateCurlCmd(id) {
    if (!els.cmdCurlArg) return;
    const s = findSkill(id);
    const url = s ? s.rawUrl : `https://raw.githubusercontent.com/${state.githubRepo}/${state.githubBranch}/skills/${id}/SKILL.md`;
    els.cmdCurlArg.textContent = url;
  }

  function updateAgent(id) {
    const s = findSkill(id);
    if (!s) return;
    if (els.agentName) els.agentName.textContent = s.name || s.id;
    if (els.agentUrl)  els.agentUrl.textContent  = s.rawUrl;
    if (els.agentBlob) {
      els.agentBlob.href = s.blobUrl;
      els.agentBlob.textContent = (s.blobUrl || "").replace(/^https?:\/\//, "");
    }
  }

  // ---------------------------------------------------------------
  // ZIP 下载弹框: nav 入口 + Esc / 背景点击关闭 + 异步拉 version.txt
  // 关键词: zip download modal binding, password disclosure, version.txt fetch
  // ---------------------------------------------------------------
  // OSS / CDN 公开常量, 与 .github/workflows/upload-hack-skills.yml 保持一致
  // 关键词: ZIP_PASSWORD, OSS root, CDN domains
  const ZIP_VERSION_URL_PRIMARY = "https://oss-qn.yaklang.com/hack-skills/latest/version.txt";
  const ZIP_VERSION_URL_FALLBACK = "https://aliyun-oss.yaklang.com/hack-skills/latest/version.txt";

  let zipVersionFetched = false;

  function openZipModal() {
    if (!els.zipModal) return;
    els.zipModal.hidden = false;
    document.documentElement.style.overflow = "hidden";
    fetchZipVersion();
    setTimeout(() => {
      const btn = els.modalZipDlPrimary || els.zipModalClose;
      if (btn && typeof btn.focus === "function") btn.focus();
    }, 50);
  }

  function closeZipModal() {
    if (!els.zipModal) return;
    els.zipModal.hidden = true;
    document.documentElement.style.overflow = "";
    if (els.navDownload && typeof els.navDownload.focus === "function") {
      els.navDownload.focus();
    }
  }

  function fetchZipVersion() {
    if (zipVersionFetched || !els.modalZipVersion) return;
    zipVersionFetched = true;
    const fallbackVersion = (state.raw && state.raw.version) || "unknown";
    const tryFetch = (url) => fetch(url, { mode: "cors", cache: "no-cache" })
      .then(r => { if (!r.ok) throw new Error("HTTP " + r.status); return r.text(); })
      .then(t => t.trim());
    tryFetch(ZIP_VERSION_URL_PRIMARY)
      .catch(() => tryFetch(ZIP_VERSION_URL_FALLBACK))
      .then(v => { els.modalZipVersion.textContent = v || fallbackVersion; })
      .catch((err) => {
        log_warn("zip version fetch failed, falling back to bundle manifest:", err);
        els.modalZipVersion.textContent = fallbackVersion;
      });
  }

  // log_warn keeps the surface area small (avoid leaking errors when CORS forbids OSS)
  function log_warn(...args) {
    try { console && console.warn && console.warn("[hackskills/zip-modal]", ...args); }
    catch (e) { /* swallow */ }
  }

  function bindZipModal() {
    if (els.navDownload) {
      els.navDownload.addEventListener("click", (e) => {
        e.preventDefault();
        openZipModal();
      });
    }
    if (els.zipModalClose) {
      els.zipModalClose.addEventListener("click", closeZipModal);
    }
    if (els.zipModal) {
      els.zipModal.addEventListener("click", (e) => {
        if (e.target === els.zipModal) closeZipModal();
      });
    }
    document.addEventListener("keydown", (e) => {
      if (e.key === "Escape" && els.zipModal && !els.zipModal.hidden) {
        closeZipModal();
      }
    });

    // Hash deep-link: 当 URL 中带 #download 时, 直接打开弹框
    // 关键词: deep link to zip modal, hash trigger
    const tryHashOpen = () => {
      if (window.location.hash === "#download" && els.zipModal && els.zipModal.hidden) {
        openZipModal();
      }
    };
    window.addEventListener("hashchange", tryHashOpen);
    tryHashOpen();
  }

})();
