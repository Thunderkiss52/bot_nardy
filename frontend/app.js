(function () {
  const ui = {
    gameType: document.getElementById("gameType"),
    botSide: document.getElementById("botSide"),
    opponent: document.getElementById("opponent"),
    thinkTime: document.getElementById("thinkTime"),
    showTop3: document.getElementById("showTop3"),
    seed: document.getElementById("seed"),
    d1: document.getElementById("d1"),
    d2: document.getElementById("d2"),
    lineIndex: document.getElementById("lineIndex"),
    startBtn: document.getElementById("startBtn"),
    applyBtn: document.getElementById("applyBtn"),
    legalBtn: document.getElementById("legalBtn"),
    undoBtn: document.getElementById("undoBtn"),
    exportBtn: document.getElementById("exportBtn"),
    status: document.getElementById("status"),
    board: document.getElementById("board"),
    barInfo: document.getElementById("barInfo"),
    offInfo: document.getElementById("offInfo"),
    top3: document.getElementById("top3"),
    analysis: document.getElementById("analysis"),
    snapshot: document.getElementById("snapshot"),
    moveLog: document.getElementById("moveLog"),
  };

  const state = {
    game: null,
    legalLines: [],
    log: [],
  };

  const topRow = [13, 14, 15, 16, 17, 18, null, 19, 20, 21, 22, 23, 24];
  const bottomRow = [12, 11, 10, 9, 8, 7, null, 6, 5, 4, 3, 2, 1];

  async function call(method, ...args) {
    const api = window.go && window.go.desktop && window.go.desktop.API;
    if (!api || typeof api[method] !== "function") {
      throw new Error("Wails API недоступен. Соберите приложение с тегом '-tags wails'.");
    }
    return api[method](...args);
  }

  function setStatus(text, isError = false) {
    ui.status.textContent = text;
    ui.status.style.color = isError ? "#9d2014" : "#164642";
  }

  function safeInt(value, fallback) {
    const n = Number.parseInt(value, 10);
    return Number.isFinite(n) ? n : fallback;
  }

  function formatLine(line) {
    if (!line || !Array.isArray(line.moves) || line.moves.length === 0) {
      return "пас";
    }
    return line.moves.map((m) => `${m.from}/${m.to}(${m.die})`).join(" ");
  }

  function ownerClass(owner) {
    if (owner === 1) return "white";
    if (owner === 2) return "black";
    return "";
  }

  function makeChecker(color) {
    const el = document.createElement("div");
    el.className = `checker ${color}`;
    return el;
  }

  function makeStackCount(count) {
    const countBadge = document.createElement("div");
    countBadge.className = "stackCount";
    countBadge.textContent = String(count);
    return countBadge;
  }

  function drawStack(container, color, count, maxVisible) {
    if (!color || count <= 0) {
      return;
    }
    const visible = Math.min(count, maxVisible);
    for (let i = 0; i < visible; i++) {
      container.appendChild(makeChecker(color));
    }
    if (count > visible) {
      container.appendChild(makeStackCount(count));
    }
  }

  function createPoint(pointIdx, rowType, colIndex) {
    const pointState = state.game.points[pointIdx] || { owner: 0, count: 0 };

    const point = document.createElement("div");
    const tone = colIndex % 2 === 0 ? "tone-a" : "tone-b";
    point.className = `point ${rowType} ${tone}`;
    point.style.gridColumn = String(colIndex + 1);
    point.style.gridRow = rowType === "top" ? "1" : "2";

    const tri = document.createElement("div");
    tri.className = "triangle";
    point.appendChild(tri);

    const stack = document.createElement("div");
    stack.className = "checkerStack";
    drawStack(stack, ownerClass(pointState.owner), pointState.count, 6);
    point.appendChild(stack);

    const label = document.createElement("span");
    label.className = "label";
    label.textContent = String(pointIdx);
    point.appendChild(label);

    return point;
  }

  function createBar() {
    const bar = document.createElement("div");
    bar.className = "barColumn";

    const blackCount = (state.game.bar && state.game.bar[1]) || 0;
    const whiteCount = (state.game.bar && state.game.bar[0]) || 0;

    const topLabel = document.createElement("div");
    topLabel.className = "barLabel top";
    topLabel.textContent = `БАР Ч:${blackCount}`;
    bar.appendChild(topLabel);

    const topStack = document.createElement("div");
    topStack.className = "barStack top";
    drawStack(topStack, "black", blackCount, 4);
    bar.appendChild(topStack);

    const bottomLabel = document.createElement("div");
    bottomLabel.className = "barLabel bottom";
    bottomLabel.textContent = `БАР Б:${whiteCount}`;
    bar.appendChild(bottomLabel);

    const bottomStack = document.createElement("div");
    bottomStack.className = "barStack bottom";
    drawStack(bottomStack, "white", whiteCount, 4);
    bar.appendChild(bottomStack);

    return bar;
  }

  function refreshSnapshot() {
    if (!state.game) {
      ui.snapshot.textContent = "Партия не начата.";
      return;
    }
    ui.snapshot.textContent = JSON.stringify(
      {
        тип_игры: state.game.game_type,
        ход: state.game.turn,
        номер_хода: state.game.meta ? state.game.meta.move_number : 0,
        вынос: state.game.off,
        бар: state.game.bar,
      },
      null,
      2
    );
  }

  function renderBoard() {
    ui.board.innerHTML = "";
    if (!state.game || !state.game.points) {
      return;
    }

    const bar = state.game.bar || [0, 0];
    const off = state.game.off || [0, 0];
    ui.barInfo.textContent = `Бар: Б ${bar[0] || 0} | Ч ${bar[1] || 0}`;
    ui.offInfo.textContent = `Вынос: Б ${off[0] || 0} | Ч ${off[1] || 0}`;

    ui.board.appendChild(createBar());

    topRow.forEach((pointIdx, colIdx) => {
      if (pointIdx == null) {
        return;
      }
      ui.board.appendChild(createPoint(pointIdx, "top", colIdx));
    });

    bottomRow.forEach((pointIdx, colIdx) => {
      if (pointIdx == null) {
        return;
      }
      ui.board.appendChild(createPoint(pointIdx, "bottom", colIdx));
    });
  }

  function renderTop3(decision) {
    ui.top3.innerHTML = "";
    if (!decision || !Array.isArray(decision.top3) || !decision.top3.length || !ui.showTop3.checked) {
      const li = document.createElement("li");
      li.textContent = "Нет вариантов Top-3 для этого хода.";
      ui.top3.appendChild(li);
      return;
    }
    decision.top3.forEach((ev, i) => {
      const li = document.createElement("li");
      li.textContent = `#${i + 1} ${formatLine(ev.line)} | WinProb=${(ev.winprob || 0).toFixed(3)} | сим=${ev.sims || 0}`;
      ui.top3.appendChild(li);
    });
  }

  function renderAnalysis(data) {
    if (!data) {
      ui.analysis.textContent = "Пока нет анализа.";
      return;
    }
    ui.analysis.textContent = `Категория: ${translateCategory(data.category)}\nДельта: ${Number(data.delta || 0).toFixed(4)}\nЛучший: ${formatLine(data.best_line)}\nЛучший WinProb: ${Number(data.best_winprob || 0).toFixed(3)}`;
  }

  function translateCategory(category) {
    switch (category) {
      case "exact":
        return "Точно";
      case "inaccuracy":
        return "Неточность";
      case "mistake":
        return "Ошибка";
      case "blunder":
        return "Грубая ошибка";
      default:
        return category || "";
    }
  }

  function pushLog(entry) {
    state.log.unshift(entry);
    if (state.log.length > 160) {
      state.log = state.log.slice(0, 160);
    }
    ui.moveLog.innerHTML = "";
    state.log.forEach((x) => {
      const li = document.createElement("li");
      li.textContent = x;
      ui.moveLog.appendChild(li);
    });
  }

  function renderLegalLines(lines) {
    state.legalLines = lines || [];
    ui.lineIndex.innerHTML = "";
    if (state.legalLines.length === 0) {
      const op = document.createElement("option");
      op.value = "-1";
      op.textContent = "Нет ходов";
      ui.lineIndex.appendChild(op);
      return;
    }
    state.legalLines.forEach((line, i) => {
      const op = document.createElement("option");
      op.value = String(i);
      op.textContent = `${i}: ${formatLine(line)}`;
      ui.lineIndex.appendChild(op);
    });
  }

  function syncTurnStatus(resp) {
    const turn = resp && resp.state ? resp.state.turn : 0;
    const side = turn === 1 ? "Белые" : turn === 2 ? "Чёрные" : "?";
    const botTurn = !!(resp && resp.isBotTurn);
    setStatus(`Ход: ${side} (${botTurn ? "БОТ" : "ЧЕЛОВЕК"})`);
  }

  async function startGame() {
    try {
      const req = {
        gameType: ui.gameType.value,
        botSide: ui.botSide.value,
        opponent: ui.opponent.value,
        thinkTime: safeInt(ui.thinkTime.value, 8),
        showTop3: ui.showTop3.checked,
        seed: safeInt(ui.seed.value, 0),
        logPath: "moves.jsonl",
      };
      const resp = await call("StartGame", req);
      state.game = resp.state;
      renderLegalLines([]);
      renderTop3(null);
      renderAnalysis(null);
      refreshSnapshot();
      renderBoard();
      syncTurnStatus(resp);
      pushLog(`новая партия: ${req.gameType}, соперник=${req.opponent}, бот=${req.botSide}, think=${req.thinkTime}s`);
    } catch (err) {
      setStatus(String(err), true);
    }
  }

  async function loadLegal() {
    try {
      const d1 = safeInt(ui.d1.value, 0);
      const d2 = safeInt(ui.d2.value, 0);
      const lines = await call("LegalLines", d1, d2);
      renderLegalLines(lines);
      setStatus(`Легальных линий: ${lines.length}`);
    } catch (err) {
      setStatus(String(err), true);
    }
  }

  async function applyDice() {
    try {
      const d1 = safeInt(ui.d1.value, 0);
      const d2 = safeInt(ui.d2.value, 0);
      const current = await call("State");
      const isBotTurn = !!current.isBotTurn;
      let lineIndex = -1;

      if (!isBotTurn) {
        const lines = await call("LegalLines", d1, d2);
        renderLegalLines(lines);
        if (lines.length > 1 && ui.lineIndex.value === "") {
          setStatus("Выберите линию человека и нажмите «Применить» ещё раз.");
          return;
        }
        if (lines.length > 0) {
          lineIndex = safeInt(ui.lineIndex.value, 0);
          if (lineIndex < 0 || lineIndex >= lines.length) {
            setStatus("Выберите корректный индекс линии.", true);
            return;
          }
        }
      }

      const resp = await call("ApplyDice", d1, d2, lineIndex);
      state.game = resp.state;
      renderBoard();
      refreshSnapshot();
      syncTurnStatus(resp);

      if (resp.decision) {
        renderTop3(resp.decision);
        renderAnalysis(null);
        pushLog(`бот ${d1}-${d2}: ${formatLine(resp.decision.chosen_line)} p=${Number(resp.decision.chosen_prob || 0).toFixed(3)}`);
      } else {
        renderTop3(null);
        renderAnalysis(resp.analysis || null);
        if (resp.applied) {
          pushLog(`человек ${d1}-${d2}: ${formatLine(resp.applied)}`);
        } else {
          pushLog(`человек ${d1}-${d2}: пас`);
        }
      }
    } catch (err) {
      setStatus(String(err), true);
    }
  }

  async function undo() {
    try {
      const resp = await call("Undo");
      state.game = resp.state;
      renderBoard();
      refreshSnapshot();
      renderTop3(null);
      renderAnalysis(null);
      syncTurnStatus(resp);
      pushLog("отмена хода");
    } catch (err) {
      setStatus(String(err), true);
    }
  }

  async function exportState() {
    try {
      const path = prompt("Путь для экспорта состояния", "state.json");
      if (!path) {
        return;
      }
      await call("Export", path);
      setStatus(`Экспортировано в ${path}`);
      pushLog(`экспорт: ${path}`);
    } catch (err) {
      setStatus(String(err), true);
    }
  }

  ui.startBtn.addEventListener("click", startGame);
  ui.applyBtn.addEventListener("click", applyDice);
  ui.legalBtn.addEventListener("click", loadLegal);
  ui.undoBtn.addEventListener("click", undo);
  ui.exportBtn.addEventListener("click", exportState);

  setStatus("Готово. Начните новую партию.");
})();
