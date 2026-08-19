/* =============================================================================
   matrix.js - subtle ambient dot grid (replaces the old matrix-rain effect)
   关键词: ambient grid, dot canvas, theme aware, low-cost background
   设计原则:
     - 极克制: 仅作为细微的纹理感, 不抢主体内容
     - 跟随主题: dark / light 用不同 dot color
     - 不动: 不再做帧动画, 只在 resize / theme 变化时重绘 (零持续 CPU)
     - 仅纯背景, 不做任何交互
   ============================================================================= */
(function () {
  "use strict";

  const canvas = document.getElementById("ambient-canvas");
  if (!canvas) return;
  const ctx = canvas.getContext("2d", { alpha: true });

  const SPACING = 28;
  const DOT_R = 0.9;

  function dotColor() {
    const cs = getComputedStyle(document.documentElement);
    const v = cs.getPropertyValue("--grid-dot").trim();
    return v || "rgba(0,0,0,0.06)";
  }

  let dpr = 1;

  function resize() {
    dpr = Math.min(window.devicePixelRatio || 1, 2);
    const w = window.innerWidth;
    const h = window.innerHeight;
    canvas.width  = Math.floor(w * dpr);
    canvas.height = Math.floor(h * dpr);
    canvas.style.width  = w + "px";
    canvas.style.height = h + "px";
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    draw();
  }

  function draw() {
    const w = canvas.width / dpr;
    const h = canvas.height / dpr;
    ctx.clearRect(0, 0, w, h);
    ctx.fillStyle = dotColor();

    // 渲染整页范围的 dot grid (含视口外一点点 padding 防止滚动闪烁)
    // 关键词: dot grid pattern, viewport tiling
    for (let y = SPACING / 2; y < h; y += SPACING) {
      for (let x = SPACING / 2; x < w; x += SPACING) {
        ctx.beginPath();
        ctx.arc(x, y, DOT_R, 0, Math.PI * 2);
        ctx.fill();
      }
    }
  }

  let resizeTimer = null;
  window.addEventListener("resize", function () {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(resize, 120);
  });

  // 主题变化时重绘 (effects.js 通过 data-theme 属性切换主题)
  // 关键词: theme observer, MutationObserver attribute
  const observer = new MutationObserver((records) => {
    for (const r of records) {
      if (r.type === "attributes" && r.attributeName === "data-theme") {
        draw();
      }
    }
  });
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ["data-theme"] });

  resize();
})();
