/* =============================================================================
   search.js - Pure-frontend weighted fuzzy search engine
   关键词: fuzzy match, weighted scoring, field qualifier, no external deps
   字段限定语法:
     category:auth  / cat:auth   - 限定分类 (按 id 或前缀)
     tier:deep      / tier:p2    - 限定 tier (master/category/deep)
     lines:>200     / lines:<50  - 行数过滤 (>=, <=, >, <, =)
     id:xss-cross   / name:xss   - 限定 id / name
   其余 token 全文 fuzzy 评分
   ============================================================================= */
(function () {
  "use strict";

  // 加权字段
  // 关键词: weighted scoring fields
  const WEIGHT_NAME        = 5.0;
  const WEIGHT_ID          = 4.0;
  const WEIGHT_DESCRIPTION = 2.0;
  const WEIGHT_CATEGORY    = 3.0;
  const WEIGHT_EXTRA       = 1.0;

  // tier 别名 (P0/P1/P2 风格 -> 实际 tier)
  // 关键词: tier alias mapping
  const TIER_ALIASES = {
    "p0": "master",
    "p1": "category",
    "p2": "deep",
    "router": "category",
    "topic": "deep"
  };

  // ---------------------------------------------------------------
  // 标记化与查询解析
  // 关键词: query tokenization, qualifier parsing
  // ---------------------------------------------------------------
  function parseQuery(raw) {
    const result = {
      raw: raw || "",
      free: [],          // 自由 token (用于 fuzzy)
      category: null,    // string | null
      tier: null,        // string | null
      idFilter: null,
      nameFilter: null,
      lines: null        // { op: ">=", value: number }
    };
    if (!raw || !raw.trim()) return result;

    const tokens = raw.trim().toLowerCase().split(/\s+/);
    for (const tok of tokens) {
      const colonIdx = tok.indexOf(":");
      if (colonIdx > 0) {
        const key = tok.slice(0, colonIdx);
        const val = tok.slice(colonIdx + 1);
        if (!val) continue;

        if (key === "category" || key === "cat") {
          result.category = val;
          continue;
        }
        if (key === "tier") {
          result.tier = TIER_ALIASES[val] || val;
          continue;
        }
        if (key === "id") {
          result.idFilter = val;
          continue;
        }
        if (key === "name") {
          result.nameFilter = val;
          continue;
        }
        if (key === "lines" || key === "l") {
          // op + number
          const m = val.match(/^(>=|<=|>|<|=)?\s*(\d+)$/);
          if (m) {
            result.lines = { op: m[1] || "=", value: parseInt(m[2], 10) };
          }
          continue;
        }
        // 未识别 qualifier, 当作 free 字符串
        result.free.push(tok);
      } else {
        result.free.push(tok);
      }
    }
    return result;
  }

  // 子串/子序列模糊评分
  // 关键词: substring score, subsequence match, position weighting
  // allowSubseq=true 时启用子序列模糊 (仅适合短字段如 name/id, 长字段会过度匹配)
  function tokenScore(needle, haystack, allowSubseq) {
    if (!needle) return 0;
    if (!haystack) return 0;
    if (haystack === needle) return 1.0;
    const idx = haystack.indexOf(needle);
    if (idx === 0) return 0.92;
    if (idx > 0) {
      // 越靠前分数越高
      return 0.7 - Math.min(idx, 60) / 200;
    }
    if (!allowSubseq) return 0;
    if (needle.length < 3) return 0;
    // 子序列匹配 (仅在短字段且 needle 长度 >=3 时启用)
    // 关键词: subsequence fuzzy, length guard
    let hi = 0;
    for (let i = 0; i < haystack.length && hi < needle.length; i++) {
      if (haystack[i] === needle[hi]) hi++;
    }
    if (hi === needle.length) {
      return 0.30 + 0.40 * (needle.length / haystack.length);
    }
    return 0;
  }

  function fieldScore(needle, haystack, allowSubseq) {
    if (!haystack) return 0;
    return tokenScore(needle, haystack.toLowerCase(), allowSubseq === true);
  }

  function checkLineFilter(filter, value) {
    if (!filter) return true;
    switch (filter.op) {
      case ">":  return value >  filter.value;
      case "<":  return value <  filter.value;
      case ">=": return value >= filter.value;
      case "<=": return value <= filter.value;
      case "=":  return value === filter.value;
      default:   return true;
    }
  }

  // ---------------------------------------------------------------
  // 主搜索: 输入 skills 数组 + query string, 返回排序后的过滤结果
  // 关键词: main search function, weighted aggregation
  // ---------------------------------------------------------------
  function search(skills, queryStr) {
    const q = parseQuery(queryStr);
    const out = [];

    for (const s of skills) {
      // 1. 硬过滤 (qualifier)
      if (q.tier && s.tier !== q.tier) continue;

      if (q.category) {
        // 支持 id 精确或前缀, 例如 category:auth 命中 auth
        const cid = (s.category || "").toLowerCase();
        if (cid !== q.category && !cid.startsWith(q.category) && !q.category.startsWith(cid)) {
          continue;
        }
      }

      if (q.idFilter && !(s.id || "").toLowerCase().includes(q.idFilter)) continue;
      if (q.nameFilter && !(s.name || "").toLowerCase().includes(q.nameFilter)) continue;
      if (!checkLineFilter(q.lines, s.skillLines | 0)) continue;

      // 2. fuzzy 评分
      let score = 0;
      let allMatch = true;

      if (q.free.length === 0) {
        // 没有自由 token, 所有通过硬过滤的都给个基线分
        score = 1;
      } else {
        for (const tok of q.free) {
          // 短字段允许子序列, 长描述仅子串 (避免长文本误命中过多)
          // 关键词: short field subseq, long field substring only
          const sName = fieldScore(tok, s.name,        true)  * WEIGHT_NAME;
          const sId   = fieldScore(tok, s.id,          true)  * WEIGHT_ID;
          const sCat  = fieldScore(tok, s.category,    true)  * WEIGHT_CATEGORY;
          const sDesc = fieldScore(tok, s.description, false) * WEIGHT_DESCRIPTION;
          let sExtra = 0;
          if (Array.isArray(s.extraDocs)) {
            for (const d of s.extraDocs) {
              sExtra = Math.max(sExtra, fieldScore(tok, d.name, true) * WEIGHT_EXTRA);
            }
          }
          const subTotal = sName + sId + sDesc + sCat + sExtra;
          if (subTotal <= 0) {
            allMatch = false;
            break;
          }
          score += subTotal;
        }
      }

      if (!allMatch) continue;

      // tier 加权: master/category 略提权, 让路由型靠前
      // 关键词: tier preference scoring
      if (s.tier === "master")   score += 0.5;
      if (s.tier === "category") score += 0.2;

      out.push({ skill: s, score: score });
    }

    if (q.free.length > 0) {
      out.sort((a, b) => b.score - a.score);
    }
    // 没有自由 token 时, 保留原顺序 (即分类默认顺序)

    return {
      query: q,
      results: out
    };
  }

  // ---------------------------------------------------------------
  // 高亮工具: 将匹配 token 用 <mark> 包裹
  // 关键词: keyword highlight, regex escape
  // ---------------------------------------------------------------
  function escapeRegex(s) {
    return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  }
  function escapeHtml(s) {
    if (s == null) return "";
    return String(s)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  function highlight(text, tokens) {
    const html = escapeHtml(text || "");
    if (!tokens || tokens.length === 0) return html;
    // 按长度倒序避免短 token 抢先匹配
    const sorted = tokens.slice().sort((a, b) => b.length - a.length).filter(t => t && t.length > 0);
    if (sorted.length === 0) return html;
    const re = new RegExp("(" + sorted.map(escapeRegex).join("|") + ")", "gi");
    return html.replace(re, "<mark>$1</mark>");
  }

  // ---------------------------------------------------------------
  // 导出
  // 关键词: HackSkillsSearch namespace export
  // ---------------------------------------------------------------
  window.HackSkillsSearch = {
    parseQuery,
    search,
    highlight,
    escapeHtml
  };
})();
