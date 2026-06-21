/**
 * charts.js — минималистичная отрисовка графиков на <canvas> без внешних
 * библиотек. Три типа графиков: купол эффективности (с отметкой экстремума),
 * кривая роста знаний (закрашенная область — намёк на интеграл) и
 * столбцы производной (скорость роста по дням).
 */

(function (global) {
  'use strict';

  function dpr() {
    return Math.max(1, window.devicePixelRatio || 1);
  }

  function setupCanvas(canvas) {
    const ratio = dpr();
    const rect = canvas.getBoundingClientRect();
    canvas.width = Math.round(rect.width * ratio);
    canvas.height = Math.round(rect.height * ratio);
    const ctx = canvas.getContext('2d');
    ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
    return { ctx, w: rect.width, h: rect.height };
  }

  const PADDING = { top: 16, right: 16, bottom: 28, left: 36 };

  function scaleX(x, min, max, w) {
    return PADDING.left + ((x - min) / (max - min)) * (w - PADDING.left - PADDING.right);
  }
  function scaleY(y, min, max, h) {
    return h - PADDING.bottom - ((y - min) / (max - min)) * (h - PADDING.top - PADDING.bottom);
  }

  function drawAxes(ctx, w, h, xLabel, yLabels) {
    ctx.strokeStyle = 'rgba(21,23,28,0.12)';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(PADDING.left, PADDING.top);
    ctx.lineTo(PADDING.left, h - PADDING.bottom);
    ctx.lineTo(w - PADDING.right, h - PADDING.bottom);
    ctx.stroke();

    ctx.fillStyle = '#8a8d96';
    ctx.font = '11px "JetBrains Mono", monospace';
    ctx.textAlign = 'left';
    ctx.textBaseline = 'middle';
  }

  /**
   * Купол эффективности E(x) с отметкой точки экстремума.
   */
  function drawEfficiencyChart(canvas, points, optimalHours) {
    const { ctx, w, h } = setupCanvas(canvas);
    ctx.clearRect(0, 0, w, h);

    const xs = points.map(p => p.hours);
    const ys = points.map(p => p.efficiency);
    const xMin = Math.min(...xs), xMax = Math.max(...xs);
    const yMin = 0, yMax = 105;

    drawAxes(ctx, w, h);

    // ось X — подписи часов через каждые 2
    ctx.fillStyle = '#8a8d96';
    ctx.font = '10.5px "JetBrains Mono", monospace';
    ctx.textAlign = 'center';
    for (let hr = 0; hr <= xMax; hr += 2) {
      const px = scaleX(hr, xMin, xMax, w);
      ctx.fillText(String(hr), px, h - PADDING.bottom + 16);
    }

    // заливка под кривой
    ctx.beginPath();
    points.forEach((p, i) => {
      const px = scaleX(p.hours, xMin, xMax, w);
      const py = scaleY(p.efficiency, yMin, yMax, h);
      if (i === 0) ctx.moveTo(px, py); else ctx.lineTo(px, py);
    });
    ctx.lineTo(scaleX(xs[xs.length - 1], xMin, xMax, w), h - PADDING.bottom);
    ctx.lineTo(scaleX(xs[0], xMin, xMax, w), h - PADDING.bottom);
    ctx.closePath();
    const grad = ctx.createLinearGradient(0, PADDING.top, 0, h - PADDING.bottom);
    grad.addColorStop(0, 'rgba(232,163,61,0.32)');
    grad.addColorStop(1, 'rgba(232,163,61,0.02)');
    ctx.fillStyle = grad;
    ctx.fill();

    // линия кривой
    ctx.beginPath();
    points.forEach((p, i) => {
      const px = scaleX(p.hours, xMin, xMax, w);
      const py = scaleY(p.efficiency, yMin, yMax, h);
      if (i === 0) ctx.moveTo(px, py); else ctx.lineTo(px, py);
    });
    ctx.strokeStyle = '#e8a33d';
    ctx.lineWidth = 2.4;
    ctx.lineJoin = 'round';
    ctx.stroke();

    // точка экстремума
    let closest = points[0], best = Infinity;
    points.forEach(p => {
      const d = Math.abs(p.hours - optimalHours);
      if (d < best) { best = d; closest = p; }
    });
    const ox = scaleX(optimalHours, xMin, xMax, w);
    const oy = scaleY(closest.efficiency, yMin, yMax, h);

    ctx.setLineDash([3, 3]);
    ctx.strokeStyle = 'rgba(21,23,28,0.3)';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(ox, oy);
    ctx.lineTo(ox, h - PADDING.bottom);
    ctx.stroke();
    ctx.setLineDash([]);

    ctx.beginPath();
    ctx.arc(ox, oy, 5, 0, Math.PI * 2);
    ctx.fillStyle = '#15171c';
    ctx.fill();
    ctx.beginPath();
    ctx.arc(ox, oy, 5, 0, Math.PI * 2);
    ctx.strokeStyle = '#f7f5f0';
    ctx.lineWidth = 2;
    ctx.stroke();

    ctx.fillStyle = '#15171c';
    ctx.font = '600 11px "JetBrains Mono", monospace';
    ctx.textAlign = 'center';
    ctx.fillText(optimalHours.toFixed(1) + 'ч', ox, oy - 14);
  }

  /**
   * Кривая роста знаний K(t) — закрашенная площадь намекает на то, что
   * итоговое значение есть интеграл (накопленная сумма) прироста по дням.
   */
  function drawKnowledgeChart(canvas, points) {
    const { ctx, w, h } = setupCanvas(canvas);
    ctx.clearRect(0, 0, w, h);

    const xs = points.map(p => p.day);
    const xMin = 0, xMax = Math.max(...xs);
    const yMin = 0, yMax = 100;

    drawAxes(ctx, w, h);

    ctx.fillStyle = '#8a8d96';
    ctx.font = '10.5px "JetBrains Mono", monospace';
    ctx.textAlign = 'center';
    const step = Math.max(1, Math.round(xMax / 6));
    for (let d = 0; d <= xMax; d += step) {
      const px = scaleX(d, xMin, xMax, w);
      ctx.fillText(String(d), px, h - PADDING.bottom + 16);
    }

    ctx.beginPath();
    points.forEach((p, i) => {
      const px = scaleX(p.day, xMin, xMax, w);
      const py = scaleY(p.knowledge, yMin, yMax, h);
      if (i === 0) ctx.moveTo(px, py); else ctx.lineTo(px, py);
    });
    ctx.lineTo(scaleX(xMax, xMin, xMax, w), h - PADDING.bottom);
    ctx.lineTo(scaleX(0, xMin, xMax, w), h - PADDING.bottom);
    ctx.closePath();
    const grad = ctx.createLinearGradient(0, PADDING.top, 0, h - PADDING.bottom);
    grad.addColorStop(0, 'rgba(91,110,232,0.30)');
    grad.addColorStop(1, 'rgba(91,110,232,0.02)');
    ctx.fillStyle = grad;
    ctx.fill();

    ctx.beginPath();
    points.forEach((p, i) => {
      const px = scaleX(p.day, xMin, xMax, w);
      const py = scaleY(p.knowledge, yMin, yMax, h);
      if (i === 0) ctx.moveTo(px, py); else ctx.lineTo(px, py);
    });
    ctx.strokeStyle = '#5b6ee8';
    ctx.lineWidth = 2.4;
    ctx.lineJoin = 'round';
    ctx.stroke();

    // финальная точка
    const last = points[points.length - 1];
    const lx = scaleX(last.day, xMin, xMax, w);
    const ly = scaleY(last.knowledge, yMin, yMax, h);
    ctx.beginPath();
    ctx.arc(lx, ly, 4.5, 0, Math.PI * 2);
    ctx.fillStyle = '#5b6ee8';
    ctx.fill();
    ctx.strokeStyle = '#f7f5f0';
    ctx.lineWidth = 1.5;
    ctx.stroke();
  }

  /**
   * Столбчатая диаграмма производной K'(t) — скорость роста по дням.
   */
  function drawDerivativeChart(canvas, points) {
    const { ctx, w, h } = setupCanvas(canvas);
    ctx.clearRect(0, 0, w, h);

    const data = points.filter(p => p.day > 0);
    if (!data.length) return;

    const xs = data.map(p => p.day);
    const xMin = 0, xMax = Math.max(...xs);
    const yMax = Math.max(...data.map(p => p.growthRate)) * 1.15 || 1;

    drawAxes(ctx, w, h);

    ctx.fillStyle = '#8a8d96';
    ctx.font = '10.5px "JetBrains Mono", monospace';
    ctx.textAlign = 'center';
    const step = Math.max(1, Math.round(xMax / 6));
    for (let d = 0; d <= xMax; d += step) {
      const px = scaleX(d, xMin, xMax, w);
      ctx.fillText(String(d), px, h - PADDING.bottom + 16);
    }

    const barWidth = Math.max(1.5, ((w - PADDING.left - PADDING.right) / data.length) * 0.7);

    data.forEach(p => {
      const px = scaleX(p.day, xMin, xMax, w);
      const py = scaleY(p.growthRate, 0, yMax, h);
      const baseY = h - PADDING.bottom;
      ctx.fillStyle = 'rgba(95,174,114,0.65)';
      ctx.fillRect(px - barWidth / 2, py, barWidth, baseY - py);
    });
  }

  global.ExamplanCharts = {
    drawEfficiencyChart,
    drawKnowledgeChart,
    drawDerivativeChart
  };

})(window);
