/**
 * app.js — логика интерфейса: переключение вкладок, синхронизация
 * слайдеров/полей формы, отправка запроса на /api/calculate,
 * отрисовка результатов и работа с историей расчётов.
 */
(function () {
  'use strict';

  // ---------- вкладки ----------
  const navlinks = document.querySelectorAll('.navlink');
  const panels = document.querySelectorAll('[data-tab-panel]');

  navlinks.forEach(btn => {
    btn.addEventListener('click', () => {
      const tab = btn.dataset.tab;
      navlinks.forEach(b => b.classList.toggle('is-active', b === btn));
      panels.forEach(p => p.hidden = p.id !== `tab-${tab}`);
      if (tab === 'history') loadHistory();
    });
  });

  // ---------- синхронизация range <-> number ----------
  function linkRangeNumber(rangeId, numberId) {
    const range = document.getElementById(rangeId);
    const number = document.getElementById(numberId);
    range.addEventListener('input', () => { number.value = range.value; });
    number.addEventListener('input', () => {
      let v = parseFloat(number.value);
      if (isNaN(v)) return;
      const min = parseFloat(range.min), max = parseFloat(range.max);
      if (v < min) v = min;
      if (v > max) v = max;
      range.value = v;
    });
  }
  linkRangeNumber('days-range', 'days');
  linkRangeNumber('hours-range', 'hours');
  linkRangeNumber('level-range', 'level');

  // ---------- сложность (segmented control) ----------
  let difficulty = 3;
  const segWrap = document.getElementById('difficulty-seg');
  segWrap.addEventListener('click', e => {
    const btn = e.target.closest('.seg__opt');
    if (!btn) return;
    difficulty = parseInt(btn.dataset.val, 10);
    segWrap.querySelectorAll('.seg__opt').forEach(b => b.classList.toggle('is-active', b === btn));
  });

  // ---------- отправка формы ----------
  const form = document.getElementById('plan-form');
  const resultsEl = document.getElementById('results');
  const emptyState = document.getElementById('empty-state');
  const calcBtn = document.getElementById('calc-btn');

  form.addEventListener('submit', async e => {
    e.preventDefault();

    const payload = {
      subject: document.getElementById('subject').value.trim() || 'Предмет',
      daysLeft: parseInt(document.getElementById('days').value, 10),
      freeHours: parseFloat(document.getElementById('hours').value),
      difficulty: difficulty,
      currentLevel: parseFloat(document.getElementById('level').value),
      save: document.getElementById('save-plan').checked
    };

    calcBtn.disabled = true;
    calcBtn.style.opacity = '0.65';

    try {
      const res = await fetch('/api/calculate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      const data = await res.json();
      if (!res.ok) {
        alert(data.error || 'Ошибка расчёта');
        return;
      }
      renderResult(data.result);
    } catch (err) {
      alert('Не удалось связаться с сервером: ' + err.message);
    } finally {
      calcBtn.disabled = false;
      calcBtn.style.opacity = '1';
    }
  });

  function riskClass(risk) {
    if (risk === 'низкий') return 'risk-low';
    if (risk === 'средний') return 'risk-medium';
    return 'risk-high';
  }

  function renderResult(result) {
    emptyState.hidden = true;
    resultsEl.hidden = false;

    document.getElementById('stat-hours').textContent = result.optimalHours.toFixed(1);
    document.getElementById('stat-result').textContent = result.predictedResult.toFixed(0) + '%';

    const riskEl = document.getElementById('stat-risk');
    riskEl.textContent = result.burnoutRisk;
    riskEl.className = 'stat__value stat__value--risk ' + riskClass(result.burnoutRisk);
    document.getElementById('stat-risk-msg').textContent = result.burnoutMessage;

    // отрисовка после того, как контейнеры стали видимыми (нужны реальные размеры)
    requestAnimationFrame(() => {
      ExamplanCharts.drawEfficiencyChart(
        document.getElementById('chart-efficiency'),
        result.efficiencyCurve,
        result.optimalHours
      );
      ExamplanCharts.drawKnowledgeChart(
        document.getElementById('chart-knowledge'),
        result.knowledgeCurve
      );
      ExamplanCharts.drawDerivativeChart(
        document.getElementById('chart-derivative'),
        result.knowledgeCurve
      );
    });
  }

  // перерисовка графиков при ресайзе окна (canvas зависит от размеров контейнера)
  let resizeTimer;
  window.addEventListener('resize', () => {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(() => {
      if (!resultsEl.hidden && window.__lastResult) {
        renderResult(window.__lastResult);
      }
    }, 150);
  });

  const _origRender = renderResult;
  renderResult = function (result) {
    window.__lastResult = result;
    _origRender(result);
  };

  // ---------- история ----------
  const historyList = document.getElementById('history-list');
  const historyEmpty = document.getElementById('history-empty');

  async function loadHistory() {
    try {
      const res = await fetch('/api/history');
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'ошибка загрузки истории');
      renderHistory(data || []);
    } catch (err) {
      historyList.innerHTML = '';
      historyEmpty.textContent = 'Не удалось загрузить историю: ' + err.message;
      historyEmpty.hidden = false;
    }
  }

  function renderHistory(records) {
    historyList.querySelectorAll('.history-item').forEach(el => el.remove());

    if (!records.length) {
      historyEmpty.hidden = false;
      return;
    }
    historyEmpty.hidden = true;

    records.forEach(rec => {
      const el = document.createElement('div');
      el.className = 'card history-item';
      const date = new Date(rec.createdAt);
      const dateStr = date.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: '2-digit' });

      el.innerHTML = `
        <div class="history-item__main">
          <span class="history-item__subject">${escapeHtml(rec.subject)}</span>
          <span class="history-item__meta">${dateStr} · ${rec.input.daysLeft} дн. · сложность ${rec.input.difficulty}</span>
        </div>
        <div class="history-item__stats">
          <div class="history-item__stat">
            <span class="v">${rec.result.optimalHours.toFixed(1)}ч</span>
            <span class="l">оптимум</span>
          </div>
          <div class="history-item__stat">
            <span class="v">${rec.result.predictedResult.toFixed(0)}%</span>
            <span class="l">результат</span>
          </div>
        </div>
        <button class="history-item__del" title="Удалить" data-id="${rec.id}">
          <svg viewBox="0 0 16 16" width="14" height="14" fill="none"><path d="M3 4h10M6.5 4V2.5h3V4M4.5 4l.5 9.5h6L11.5 4" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
      `;
      historyList.appendChild(el);
    });

    historyList.querySelectorAll('.history-item__del').forEach(btn => {
      btn.addEventListener('click', async () => {
        const id = btn.dataset.id;
        try {
          const res = await fetch(`/api/history/delete?id=${id}`, { method: 'DELETE' });
          if (!res.ok) throw new Error('ошибка удаления');
          loadHistory();
        } catch (err) {
          alert('Не удалось удалить запись: ' + err.message);
        }
      });
    });
  }

  function escapeHtml(str) {
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
  }

})();
