/* =============================================================================
   effects.js - theme toggle (light/dark/system) + global copy + toast
   关键词: theme controller, prefers-color-scheme tracking, clipboard copy, toast feedback
   ============================================================================= */
(function () {
  "use strict";

  const THEME_KEY = "hackskills:theme";
  const root = document.documentElement;

  function getSystem() {
    return (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches)
      ? "dark"
      : "light";
  }

  function applyTheme(pref) {
    const effective = pref === "system" ? getSystem() : pref;
    root.setAttribute("data-theme", effective);
    root.setAttribute("data-theme-pref", pref);
    const btn = document.getElementById("theme-toggle");
    if (btn) {
      const labels = { light: "Theme: light", dark: "Theme: dark", system: "Theme: system (auto)" };
      btn.title = labels[pref] || labels.system;
      btn.setAttribute("aria-label", labels[pref] || labels.system);
    }
  }

  function getStored() {
    try {
      return localStorage.getItem(THEME_KEY) || "system";
    } catch (e) { return "system"; }
  }

  function setStored(v) {
    try { localStorage.setItem(THEME_KEY, v); } catch (e) {}
  }

  // Cycle: system -> light -> dark -> system
  function cycle() {
    const cur = getStored();
    const next = cur === "system" ? "light"
              : cur === "light"  ? "dark"
              :                    "system";
    setStored(next);
    applyTheme(next);
    showToast("theme: " + next);
    return next;
  }

  // ---------------------------------------------------------------
  // Toast (used by copy + theme switch)
  // 关键词: toast feedback, transient notice
  // ---------------------------------------------------------------
  let toastTimer = null;
  function showToast(msg) {
    const t = document.getElementById("toast");
    if (!t) return;
    t.textContent = msg;
    t.classList.add("toast-show");
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => t.classList.remove("toast-show"), 1400);
  }

  // ---------------------------------------------------------------
  // Generic copy: any [data-copy-target="#selector"] or [data-copy-text]
  // 关键词: clipboard copy, fallback execCommand, copied state
  // ---------------------------------------------------------------
  function copyText(text) {
    if (!text) return Promise.resolve(false);
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text).then(() => true).catch(() => fallbackCopy(text));
    }
    return Promise.resolve(fallbackCopy(text));
  }

  function fallbackCopy(text) {
    try {
      const ta = document.createElement("textarea");
      ta.value = text;
      ta.setAttribute("readonly", "");
      ta.style.position = "absolute";
      ta.style.left = "-9999px";
      document.body.appendChild(ta);
      ta.select();
      const ok = document.execCommand("copy");
      document.body.removeChild(ta);
      return ok;
    } catch (e) { return false; }
  }

  function flashCopied(btn) {
    if (!btn) return;
    btn.setAttribute("data-copied", "1");
    setTimeout(() => btn.removeAttribute("data-copied"), 1100);
  }

  function bindGlobalCopy() {
    document.addEventListener("click", (e) => {
      const btn = e.target.closest("[data-copy-target], [data-copy-text]");
      if (!btn) return;
      e.preventDefault();
      let text = btn.dataset.copyText;
      if (!text) {
        const sel = btn.dataset.copyTarget;
        const node = sel ? document.querySelector(sel) : null;
        if (node) {
          text = node.innerText || node.textContent || "";
        }
      }
      if (!text) return;
      copyText(text.trim()).then((ok) => {
        if (ok) {
          flashCopied(btn);
          showToast("copied");
        } else {
          showToast("copy failed");
        }
      });
    });
  }

  // ---------------------------------------------------------------
  // Init
  // ---------------------------------------------------------------
  applyTheme(getStored());

  // System theme change tracking (only when in system mode)
  // 关键词: prefers-color-scheme listener
  if (window.matchMedia) {
    const mql = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = () => {
      if (getStored() === "system") applyTheme("system");
    };
    if (mql.addEventListener) mql.addEventListener("change", onChange);
    else if (mql.addListener) mql.addListener(onChange);
  }

  document.addEventListener("DOMContentLoaded", function () {
    const btn = document.getElementById("theme-toggle");
    if (btn) {
      btn.addEventListener("click", cycle);
    }
    bindGlobalCopy();
  });

  // 公开 API 供 app.js / 其它脚本使用
  // 关键词: HackSkillsFx public API
  window.HackSkillsFx = {
    showToast,
    copyText,
    cycleTheme: cycle,
    getTheme: getStored,
    applyTheme: function (pref) {
      setStored(pref);
      applyTheme(pref);
    }
  };
})();
