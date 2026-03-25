(function () {
  const ui = {
    gameType: document.getElementById("gameType"),
    botSide: document.getElementById("botSide"),
    opponent: document.getElementById("opponent"),
    thinkTime: document.getElementById("thinkTime"),
    showTop3: document.getElementById("showTop3"),
    autoHumanRoll: document.getElementById("autoHumanRoll"),
    seed: document.getElementById("seed"),
    d1: document.getElementById("d1"),
    d2: document.getElementById("d2"),
    diceInput: document.getElementById("diceInput"),
    syncDiceBtn: document.getElementById("syncDiceBtn"),
    diceTray: document.getElementById("diceTray"),
    dieBtn1: document.getElementById("dieBtn1"),
    dieBtn2: document.getElementById("dieBtn2"),
    dieBtn3: document.getElementById("dieBtn3"),
    dieBtn4: document.getElementById("dieBtn4"),
    lineIndex: document.getElementById("lineIndex"),
    langSelect: document.getElementById("langSelect"),
    startBtn: document.getElementById("startBtn"),
    swapSideBtn: document.getElementById("swapSideBtn"),
    applyBtn: document.getElementById("applyBtn"),
    legalBtn: document.getElementById("legalBtn"),
    clearSelectionBtn: document.getElementById("clearSelectionBtn"),
    editBtn: document.getElementById("editBtn"),
    undoBtn: document.getElementById("undoBtn"),
    exportBtn: document.getElementById("exportBtn"),
    selfLearnBtn: document.getElementById("selfLearnBtn"),
    bgTrainingBtn: document.getElementById("bgTrainingBtn"),
    trainingStatus: document.getElementById("trainingStatus"),
    turnGuide: document.getElementById("turnGuide"),
    status: document.getElementById("status"),
    selectedPath: document.getElementById("selectedPath"),
    board: document.getElementById("board"),
    barInfo: document.getElementById("barInfo"),
    offInfo: document.getElementById("offInfo"),
    top3: document.getElementById("top3"),
    analysis: document.getElementById("analysis"),
    snapshot: document.getElementById("snapshot"),
    moveLog: document.getElementById("moveLog"),
  };

  const I18N = {
    ru: {
      "hero.title": "Движок Нард",
      "hero.subtitle": "Играй с ботом: кости находятся у самой доски, бросок делается кликом по ним, ручной ввод 4:2 доступен рядом.",
      "labels.language": "Язык",
      "labels.gameType": "Тип игры",
      "labels.botSide": "Сторона бота",
      "labels.opponent": "Соперник",
      "labels.think": "Время на ход (1-5)",
      "labels.seed": "Seed",
      "labels.seedPlaceholder": "опционально",
      "labels.showTop3": "Показывать Top-3",
      "labels.autoHumanRoll": "Автоигра ботов (только бот vs бот)",
      "labels.lineIndex": "Линия (человек)",
      "labels.diceInput": "Кубики / 4:2",
      "labels.diceHint": "Кости встроены в доску. Клик по ним бросает ход, ввод 4:2 находится рядом с доской.",
      "sections.setup": "Параметры партии",
      "sections.turn": "Ход",
      "sections.board": "Доска",
      "sections.info": "Информация",
      "sections.log": "Лог ходов",
      "sections.analysis": "Анализ",
      "buttons.newGame": "Новая партия",
      "buttons.swapSide": "Сменить сторону",
      "buttons.syncDice": "Бросить эти кости",
      "buttons.showLegal": "Показать легальные",
      "buttons.clearSelection": "Сброс выбора",
      "buttons.editCheckers": "Изменить шашки",
      "buttons.apply": "Применить ход",
      "buttons.undo": "Отменить",
      "buttons.export": "Экспорт",
      "buttons.selfLearn": "Самообучение",
      "buttons.bgTrainingStart": "Фоновый self-play",
      "buttons.bgTrainingStop": "Стоп self-play",
      "hint.click": "Для хода: кликни по костям на доске, затем кликай по подсвеченным фишкам и целям.",
      "options.short": "Короткие",
      "options.long": "Длинные",
      "options.white": "Белые",
      "options.black": "Чёрные",
      "options.human": "Человек",
      "options.bot": "Бот",
      "common.pass": "пас",
      "common.none": "Нет данных",
      "common.bar": "БАР",
      "common.off": "ВЫНОС",
      "common.noMoves": "Нет ходов",
      "common.noTop3": "Нет вариантов Top-3 для этого хода.",
      "common.noAnalysis": "Пока нет анализа.",
      "common.trainingOff": "Фоновое обучение выключено.",
      "status.ready": "Готово. Начните новую партию.",
      "status.turn": "Ход: {side} ({actor})",
      "status.legalCount": "Легальных линий: {count}",
      "status.noLegal": "Нет легальных ходов для этих кубиков.",
      "status.loadLegalFirst": "Сначала выберите кубики и нажмите «Показать легальные».",
      "status.needLine": "Выберите линию: через список или кликами по доске.",
      "status.invalidLine": "Выбран некорректный индекс линии.",
      "status.waitBot": "Сейчас ход бота.",
      "status.pickSource": "Выберите фишку (исходную точку) подсвеченного хода.",
      "status.pickDestination": "Точка {point} выбрана. Теперь укажите, куда ходить.",
      "status.pickHighlighted": "Выберите подсвеченную исходную точку.",
      "status.badDestination": "Выберите подсвеченную точку назначения.",
      "status.selectionCleared": "Выбор хода сброшен.",
      "status.editModeOn": "Режим исправления включён. Выберите шашку на доске.",
      "status.editCancelled": "Режим исправления выключен.",
      "status.editPickDestination": "Шашка {point} выбрана. Укажите новую позицию.",
      "status.editDone": "Позиция исправлена: {from} -> {to}. Можно переставлять дальше.",
      "status.editExitFirst": "Сначала выйдите из режима «Изменить шашки» этой же кнопкой.",
      "status.editNeedChecker": "Для исправления сначала выберите шашку на доске.",
      "status.editBadDestination": "Сюда нельзя переставить выбранную шашку.",
      "status.editSourceOnly": "Исходную шашку для исправления можно взять только с доски.",
      "status.lineSelected": "Линия #{index} собрана кликами. Можно применять ход.",
      "status.stepAccepted": "Добавлен шаг: {from}/{to}({die}).",
      "status.stepRemoved": "Последний шаг снят. Можно выбрать другой.",
      "status.randomReady": "Кости на доске: {d1}-{d2}.",
      "status.diceChanged": "Кубики изменились. Проверьте легальные ходы заново.",
      "status.exported": "Экспортировано в {path}",
      "status.noGame": "Сначала начните новую партию.",
      "status.selectionLocked": "Линия уже собрана. Нажмите «Применить ход» или «Сброс выбора».",
      "status.chooseLineAgain": "Выберите линию и нажмите «Применить ход» ещё раз.",
      "status.passApplied": "Ходов нет, выполнен пас.",
      "status.busy": "Подождите, выполняется ход...",
      "status.sideSwapped": "Стороны поменялись. Бот теперь играет за {side}.",
      "status.autoHumanReady": "Ваш ход: выберите подсвеченную фишку, затем точку назначения.",
      "status.autoBotMove": "Ход бота выполнен автоматически.",
      "status.autoNoMoves": "По этим кубикам ходов нет, выполнен пас.",
      "status.autoApplied": "Ход применён автоматически.",
      "status.selfLearnDone": "Самообучение завершено: примеров {examples}, эпох {epochs}, принято {accepted}.",
      "status.bgStarted": "Фоновый self-play запущен.",
      "status.bgStopped": "Фоновый self-play остановлен.",
      "guide.idle": "1) Новая партия 2) Клик по костям на доске 3) Клик по фишке и точке назначения.",
      "guide.edit": "Режим исправления: выберите шашку на доске, затем точку, бар или вынос.",
      "guide.bot": "Сейчас ход бота: он сам бросит и сходит.",
      "guide.botManual": "Сейчас ход бота: бросьте или введите кости, после этого бот сходит.",
      "guide.humanDice": "Ваш ход: кликните по костям на доске или введите 4:2.",
      "guide.humanMove": "Ваш ход: кликайте по подсвеченным точкам, затем ход применится.",
      "dice.roll": "бросок",
      "dice.used": "исп.",
      "dice.double": "дубль",
      "dice.empty": "жми",
      "path.empty": "Выбор хода: пока пусто.",
      "path.editIdle": "Исправление позиции: выберите шашку на доске.",
      "path.editSource": "Исправление позиции: {from} -> ?",
      "path.current": "Выбор хода: {path}",
      "path.ready": "Линия готова к применению.",
      "path.variants": "Осталось вариантов: {count}",
      "side.white": "Белые",
      "side.black": "Чёрные",
      "actor.bot": "БОТ",
      "actor.human": "ЧЕЛОВЕК",
      "analysis.category": "Категория",
      "analysis.delta": "Дельта",
      "analysis.best": "Лучший",
      "analysis.bestProb": "Лучший WinProb",
      "category.exact": "Точно",
      "category.inaccuracy": "Неточность",
      "category.mistake": "Ошибка",
      "category.blunder": "Грубая ошибка",
      "stats.bar": "Бар: Б {white} | Ч {black}",
      "stats.off": "Вынос: Б {white} | Ч {black}",
      "bar.white": "БАР Б:{count}",
      "bar.black": "БАР Ч:{count}",
      "line.option": "#{index}: {line}",
      "log.newGame": "новая партия: {game}, соперник={opponent}, бот={bot}, think={think}s",
      "log.botMove": "бот {d1}-{d2}: {line} p={prob}",
      "log.humanMove": "человек {d1}-{d2}: {line}",
      "log.pass": "человек {d1}-{d2}: пас",
      "log.undo": "отмена хода",
      "log.swapSide": "смена стороны: бот теперь {side}",
      "log.export": "экспорт: {path}",
      "log.editChecker": "исправление позиции: {from} -> {to}",
      "prompt.export": "Путь для экспорта состояния",
      "errors.wails": "Wails API недоступен. Соберите приложение с тегом '-tags wails'.",
      "errors.dice": "Кубики должны быть в диапазоне 1..6. Можно вводить 42, 4 2 или 4:2.",
      "errors.lineRequired": "Нужно выбрать линию для хода человека.",
      "errors.turnBot": "Сейчас ход бота.",
      "errors.turnHuman": "Сейчас ход человека.",
      "errors.undo": "Нет хода для отмены.",
      "errors.exportPath": "Путь экспорта пустой.",
      "errors.outOfRange": "Индекс линии вне диапазона.",
      "errors.illegalAnalysis": "Невозможно проанализировать нелегальный ход.",
      "errors.training": "Обучение сейчас недоступно: {message}",
      "errors.noTrainingData": "Пока нет данных для обучения. Сыграй несколько партий или включи фоновый self-play.",
      "errors.editRange": "Точка редактирования вне диапазона.",
      "errors.editSame": "Исходная и целевая точки совпадают.",
      "errors.editEmpty": "В исходной точке нет шашки.",
      "errors.editOpponent": "Нельзя поставить шашку на занятую точку соперника.",
      "errors.generic": "Ошибка: {message}",
    },
    en: {
      "hero.title": "Nardy Engine",
      "hero.subtitle": "Play the bot: dice live next to the board, click them to roll, or enter 4:2 nearby.",
      "labels.language": "Language",
      "labels.gameType": "Game Type",
      "labels.botSide": "Bot Side",
      "labels.opponent": "Opponent",
      "labels.think": "Think Time (1-5)",
      "labels.seed": "Seed",
      "labels.seedPlaceholder": "optional",
      "labels.showTop3": "Show Top-3",
      "labels.autoHumanRoll": "Bot Auto-Play (bot vs bot only)",
      "labels.lineIndex": "Line (human)",
      "labels.diceInput": "Dice / 4:2",
      "labels.diceHint": "Dice are built into the board area. Click them to roll, or enter 4:2 beside the board.",
      "sections.setup": "Match Setup",
      "sections.turn": "Turn",
      "sections.board": "Board",
      "sections.info": "Info",
      "sections.log": "Move Log",
      "sections.analysis": "Analysis",
      "buttons.newGame": "New Game",
      "buttons.swapSide": "Swap Side",
      "buttons.syncDice": "Roll These Dice",
      "buttons.showLegal": "Show Legal",
      "buttons.clearSelection": "Clear Selection",
      "buttons.editCheckers": "Edit Checkers",
      "buttons.apply": "Apply Move",
      "buttons.undo": "Undo",
      "buttons.export": "Export",
      "buttons.selfLearn": "Self-Learn",
      "buttons.bgTrainingStart": "Background Self-Play",
      "buttons.bgTrainingStop": "Stop Self-Play",
      "hint.click": "To move: click the board dice, then click highlighted checkers and targets.",
      "options.short": "Short",
      "options.long": "Long",
      "options.white": "White",
      "options.black": "Black",
      "options.human": "Human",
      "options.bot": "Bot",
      "common.pass": "pass",
      "common.none": "No data",
      "common.bar": "BAR",
      "common.off": "OFF",
      "common.noMoves": "No moves",
      "common.noTop3": "No Top-3 variants for this turn.",
      "common.noAnalysis": "No analysis yet.",
      "common.trainingOff": "Background training is off.",
      "status.ready": "Ready. Start a new game.",
      "status.turn": "Turn: {side} ({actor})",
      "status.legalCount": "Legal lines: {count}",
      "status.noLegal": "No legal moves for these dice.",
      "status.loadLegalFirst": "Set dice and click 'Show Legal' first.",
      "status.needLine": "Choose a line: via dropdown or board clicks.",
      "status.invalidLine": "Selected line index is invalid.",
      "status.waitBot": "It is bot's turn now.",
      "status.pickSource": "Pick a checker (source point) from highlighted points.",
      "status.pickDestination": "Point {point} selected. Now choose destination.",
      "status.pickHighlighted": "Pick a highlighted source point.",
      "status.badDestination": "Pick a highlighted destination point.",
      "status.selectionCleared": "Move selection cleared.",
      "status.editModeOn": "Board edit mode is on. Pick a checker on the board.",
      "status.editCancelled": "Board edit mode is off.",
      "status.editPickDestination": "Checker {point} selected. Choose a new destination.",
      "status.editDone": "Board corrected: {from} -> {to}. You can keep editing.",
      "status.editExitFirst": "Exit board edit mode with the same button first.",
      "status.editNeedChecker": "Pick a checker on the board first.",
      "status.editBadDestination": "That destination is not allowed for the selected checker.",
      "status.editSourceOnly": "Edit mode can only start from a checker on the board.",
      "status.lineSelected": "Line #{index} built by clicks. You can apply move now.",
      "status.stepAccepted": "Step added: {from}/{to}({die}).",
      "status.stepRemoved": "Last step removed. You can choose another one.",
      "status.randomReady": "Board dice: {d1}-{d2}.",
      "status.diceChanged": "Dice changed. Re-check legal lines.",
      "status.exported": "Exported to {path}",
      "status.noGame": "Start a new game first.",
      "status.selectionLocked": "Line is already complete. Click Apply Move or Clear Selection.",
      "status.chooseLineAgain": "Choose a line and click Apply Move again.",
      "status.passApplied": "No legal moves, pass applied.",
      "status.busy": "Please wait, processing move...",
      "status.sideSwapped": "Sides swapped. Bot now plays {side}.",
      "status.autoHumanReady": "Your turn: choose a highlighted checker, then a destination.",
      "status.autoBotMove": "Bot move applied automatically.",
      "status.autoNoMoves": "No moves for these dice, pass applied.",
      "status.autoApplied": "Move applied automatically.",
      "status.selfLearnDone": "Self-learning finished: examples {examples}, epochs {epochs}, accepted {accepted}.",
      "status.bgStarted": "Background self-play started.",
      "status.bgStopped": "Background self-play stopped.",
      "guide.idle": "1) New game 2) Click the board dice 3) Click checker and destination.",
      "guide.edit": "Board edit mode: pick a checker, then a point, bar, or off.",
      "guide.bot": "Bot turn now: it will roll and play automatically.",
      "guide.botManual": "Bot turn now: roll or enter dice for the bot, then it will move.",
      "guide.humanDice": "Your turn: click the board dice or enter 4:2.",
      "guide.humanMove": "Your turn: click highlighted points; move will apply automatically.",
      "dice.roll": "roll",
      "dice.used": "used",
      "dice.double": "double",
      "dice.empty": "tap",
      "path.empty": "Move selection: empty.",
      "path.editIdle": "Board edit: pick a checker on the board.",
      "path.editSource": "Board edit: {from} -> ?",
      "path.current": "Move selection: {path}",
      "path.ready": "Line is ready to apply.",
      "path.variants": "Remaining variants: {count}",
      "side.white": "White",
      "side.black": "Black",
      "actor.bot": "BOT",
      "actor.human": "HUMAN",
      "analysis.category": "Category",
      "analysis.delta": "Delta",
      "analysis.best": "Best",
      "analysis.bestProb": "Best WinProb",
      "category.exact": "Exact",
      "category.inaccuracy": "Inaccuracy",
      "category.mistake": "Mistake",
      "category.blunder": "Blunder",
      "stats.bar": "Bar: W {white} | B {black}",
      "stats.off": "Off: W {white} | B {black}",
      "bar.white": "BAR W:{count}",
      "bar.black": "BAR B:{count}",
      "line.option": "#{index}: {line}",
      "log.newGame": "new game: {game}, opponent={opponent}, bot={bot}, think={think}s",
      "log.botMove": "bot {d1}-{d2}: {line} p={prob}",
      "log.humanMove": "human {d1}-{d2}: {line}",
      "log.pass": "human {d1}-{d2}: pass",
      "log.undo": "undo move",
      "log.swapSide": "side swapped: bot is now {side}",
      "log.export": "export: {path}",
      "log.editChecker": "board edit: {from} -> {to}",
      "prompt.export": "Export path",
      "errors.wails": "Wails API unavailable. Build app with '-tags wails'.",
      "errors.dice": "Dice must be in range 1..6. You can enter 42, 4 2, or 4:2.",
      "errors.lineRequired": "A line is required for human turn.",
      "errors.turnBot": "Current turn belongs to bot.",
      "errors.turnHuman": "Current turn belongs to human.",
      "errors.undo": "Nothing to undo.",
      "errors.exportPath": "Export path is empty.",
      "errors.outOfRange": "Line index is out of range.",
      "errors.illegalAnalysis": "Cannot analyze an illegal line.",
      "errors.training": "Training is unavailable right now: {message}",
      "errors.noTrainingData": "No training data yet. Play a few games or enable background self-play first.",
      "errors.editRange": "Edit point is out of range.",
      "errors.editSame": "Source and destination are the same.",
      "errors.editEmpty": "The source point is empty.",
      "errors.editOpponent": "Cannot place a checker on an opponent point.",
      "errors.generic": "Error: {message}",
    },
  };

  const state = {
    game: null,
    legalLines: [],
    log: [],
    isBotTurn: false,
    busy: false,
    diceRequestSeq: 0,
    lang: "ru",
    click: {
      d1: 0,
      d2: 0,
      path: [],
      candidates: [],
      selectedFrom: null,
      selectableFrom: new Set(),
      selectableTo: new Set(),
    },
    animation: {
      queue: [],
      playing: false,
    },
    training: {
      running: false,
      pollHandle: null,
      status: null,
    },
    edit: {
      enabled: false,
      source: null,
    },
  };

  const topRow = [13, 14, 15, 16, 17, 18, null, 19, 20, 21, 22, 23, 24];
  const bottomRow = [12, 11, 10, 9, 8, 7, null, 6, 5, 4, 3, 2, 1];

  async function call(method, ...args) {
    const api = window.go && window.go.desktop && window.go.desktop.API;
    if (!api || typeof api[method] !== "function") {
      throw new Error(t("errors.wails"));
    }
    return api[method](...args);
  }

  function t(key, vars = {}) {
    const dict = I18N[state.lang] || I18N.ru;
    const fallback = I18N.ru;
    let template = dict[key] || fallback[key] || key;
    Object.entries(vars).forEach(([name, value]) => {
      template = template.replaceAll(`{${name}}`, String(value));
    });
    return template;
  }

  function renderTrainingStatus() {
    if (!ui.trainingStatus) {
      return;
    }
    const status = state.training.status;
    if (!status || !status.running) {
      ui.trainingStatus.textContent = t("common.trainingOff");
      if (ui.bgTrainingBtn) {
        ui.bgTrainingBtn.textContent = t("buttons.bgTrainingStart");
      }
      return;
    }
    const parts = [
      `self-play: games ${status.games || 0}`,
      `examples ${status.examples || 0}`,
      `workers ${status.workers || 0}`,
    ];
    if (status.last_error) {
      parts.push(`error ${status.last_error}`);
    }
    ui.trainingStatus.textContent = parts.join(" | ");
    if (ui.bgTrainingBtn) {
      ui.bgTrainingBtn.textContent = t("buttons.bgTrainingStop");
    }
  }

  function stopTrainingPoll() {
    if (state.training.pollHandle) {
      window.clearTimeout(state.training.pollHandle);
      state.training.pollHandle = null;
    }
  }

  function applyBackgroundTrainingStatus(status) {
    state.training.status = status || null;
    state.training.running = !!(status && status.running);
    renderTrainingStatus();
    stopTrainingPoll();
    if (state.training.running) {
      state.training.pollHandle = window.setTimeout(refreshBackgroundTrainingStatus, 3000);
    }
  }

  async function refreshBackgroundTrainingStatus() {
    try {
      const resp = await call("BackgroundTrainingStatus");
      applyBackgroundTrainingStatus(resp && resp.status ? resp.status : null);
    } catch (_) {
      stopTrainingPoll();
    }
  }

  async function selfLearn() {
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    setBusy(true);
    try {
      const resp = await call("SelfLearn", 12);
      const result = resp && resp.result ? resp.result : null;
      if (!result) {
        throw new Error("empty self-learn result");
      }
      const accepted = result.accepted ? "yes" : "no";
      setStatus(t("status.selfLearnDone", { examples: result.examples || 0, epochs: result.epochs || 0, accepted }));
      pushLog(`self-learn: examples=${result.examples || 0}, accepted=${accepted}, model=${result.model_name || "n/a"}`);
      await refreshBackgroundTrainingStatus();
    } catch (err) {
      const message = String(err || "");
      if (message.toLowerCase().includes("no training examples found")) {
        setStatus(t("errors.noTrainingData"), true);
      } else {
        setStatus(t("errors.training", { message }), true);
      }
    } finally {
      setBusy(false);
      updateTurnGuide();
    }
  }

  async function toggleBackgroundTraining() {
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    try {
      const method = state.training.running ? "StopBackgroundTraining" : "StartBackgroundTraining";
      const resp = await call(method);
      applyBackgroundTrainingStatus(resp && resp.status ? resp.status : null);
      setStatus(state.training.running ? t("status.bgStarted") : t("status.bgStopped"));
      pushLog(state.training.running ? "background self-play: started" : "background self-play: stopped");
    } catch (err) {
      setStatus(t("errors.training", { message: String(err) }), true);
    }
  }

  function safeInt(value, fallback) {
    const n = Number.parseInt(value, 10);
    return Number.isFinite(n) ? n : fallback;
  }

  function parseDiceText(text) {
    const digits = String(text || "").replace(/\D/g, "");
    if (digits.length !== 2) {
      return null;
    }
    const d1 = Number(digits[0]);
    const d2 = Number(digits[1]);
    if (d1 < 1 || d1 > 6 || d2 < 1 || d2 > 6) {
      return null;
    }
    return { d1, d2 };
  }

  function currentDice() {
    return {
      d1: safeInt(ui.d1.value, 0),
      d2: safeInt(ui.d2.value, 0),
    };
  }

  function syncDiceInputText() {
    if (!ui.diceInput) {
      return;
    }
    const { d1, d2 } = currentDice();
    ui.diceInput.value = d1 >= 1 && d1 <= 6 && d2 >= 1 && d2 <= 6 ? `${d1}:${d2}` : "";
  }

  function buildDiceSlots() {
    const { d1, d2 } = currentDice();
    if (d1 < 1 || d1 > 6 || d2 < 1 || d2 > 6) {
      return [
        { value: "?", used: false, hidden: false, metaKey: "dice.empty" },
        { value: "?", used: false, hidden: false, metaKey: "dice.empty" },
        { value: "", used: false, hidden: true, metaKey: "dice.empty" },
        { value: "", used: false, hidden: true, metaKey: "dice.empty" },
      ];
    }

    const values = d1 === d2 ? [d1, d1, d1, d1] : [d1, d2];
    const slots = values.map((value) => ({
      value,
      used: false,
      hidden: false,
      metaKey: d1 === d2 ? "dice.double" : "dice.roll",
    }));
    while (slots.length < 4) {
      slots.push({ value: "", used: false, hidden: true, metaKey: "dice.empty" });
    }
    for (const mv of state.click.path) {
      const idx = slots.findIndex((slot) => !slot.hidden && !slot.used && Number(slot.value) === mv.die);
      if (idx >= 0) {
        slots[idx].used = true;
        slots[idx].metaKey = "dice.used";
      }
    }
    return slots;
  }

  function pipMarkup(value) {
    const layouts = {
      1: [0, 0, 0, 0, 1, 0, 0, 0, 0],
      2: [1, 0, 0, 0, 0, 0, 0, 0, 1],
      3: [1, 0, 0, 0, 1, 0, 0, 0, 1],
      4: [1, 0, 1, 0, 0, 0, 1, 0, 1],
      5: [1, 0, 1, 0, 1, 0, 1, 0, 1],
      6: [1, 0, 1, 1, 0, 1, 1, 0, 1],
    };
    const cells = layouts[value] || [0, 0, 0, 0, 0, 0, 0, 0, 0];
    return cells.map((filled) => `<span class="pip${filled ? "" : " hidden"}"></span>`).join("");
  }

  function renderDiceTray() {
    if (!ui.dieBtn1) {
      return;
    }
    const buttons = [ui.dieBtn1, ui.dieBtn2, ui.dieBtn3, ui.dieBtn4];
    const slots = buildDiceSlots();
    buttons.forEach((btn, index) => {
      const slot = slots[index];
      if (!slot || slot.hidden) {
        btn.className = "dieFace hidden";
        btn.disabled = true;
        btn.innerHTML = "";
        return;
      }
      btn.className = `dieFace${slot.used ? " used" : ""} rollable`;
      btn.disabled = !!state.busy;
      const valueMarkup = Number.isFinite(Number(slot.value)) ? pipMarkup(Number(slot.value)) : slot.value;
      btn.innerHTML = `<span class="dieValue">${valueMarkup}</span><span class="dieMeta">${t(slot.metaKey)}</span>`;
    });
  }

  async function setDiceValues(d1, d2, triggerReload = true) {
    ui.d1.value = String(d1);
    ui.d2.value = String(d2);
    syncDiceInputText();
    renderDiceTray();
    if (triggerReload) {
      await onDiceChanged();
    }
  }

  function setStatus(text, isError = false) {
    ui.status.textContent = text;
    ui.status.classList.toggle("error", !!isError);
  }

  function formatPoint(point) {
    if (point === 0) {
      return t("common.off");
    }
    return String(point);
  }

  function formatMove(mv) {
    if (!mv) {
      return "";
    }
    return `${formatPoint(mv.from)}/${formatPoint(mv.to)}(${mv.die})`;
  }

  function formatLine(line) {
    if (!line || !Array.isArray(line.moves) || line.moves.length === 0) {
      return t("common.pass");
    }
    return line.moves.map((m) => formatMove(m)).join(" ");
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

  function checkerColorClassForPlayer(player) {
    return ownerClass(player);
  }

  function makeStackCount(count) {
    const countBadge = document.createElement("div");
    countBadge.className = "stackCount";
    countBadge.textContent = String(count);
    return countBadge;
  }

  function drawStack(container, color, count, direction, maxVisible = 15) {
    if (!color || count <= 0) {
      return;
    }
    const visible = Math.min(count, maxVisible);
    const compact = visible > 7;
    const step = compact ? 12 : 28;

    for (let i = 0; i < visible; i++) {
      const checker = makeChecker(color);
      if (compact) {
        checker.classList.add("compact");
      }
      const offset = i * step;
      if (direction === "top") {
        checker.style.top = `${offset}px`;
      } else {
        checker.style.bottom = `${offset}px`;
      }
      container.appendChild(checker);
    }

    if (count > visible) {
      container.appendChild(makeStackCount(count));
    }
  }

  function pointIsInPath(pointIdx) {
    return state.click.path.some((mv) => mv.from === pointIdx || mv.to === pointIdx);
  }

  function cloneGameStateForPreview(src) {
    return {
      ...src,
      points: src.points.map((point) => ({ owner: point.owner, count: point.count })),
      off: src.off.slice(),
      bar: src.bar.slice(),
      meta: { ...src.meta },
    };
  }

  function applyPreviewLine(baseState, moves) {
    const preview = cloneGameStateForPreview(baseState);
    const player = preview.turn;
    const opp = player === 1 ? 2 : 1;
    const playerIdx = player === 1 ? 0 : 1;
    const oppIdx = opp === 1 ? 0 : 1;

    for (const mv of moves) {
      if (mv.from === 0) {
        preview.bar[playerIdx] = Math.max(0, (preview.bar[playerIdx] || 0) - 1);
      } else {
        const src = preview.points[mv.from];
        src.count = Math.max(0, (src.count || 0) - 1);
        if (src.count === 0) {
          src.owner = 0;
        }
      }

      if (mv.to === 0) {
        preview.off[playerIdx] = (preview.off[playerIdx] || 0) + 1;
        continue;
      }

      const dst = preview.points[mv.to];
      if (preview.game_type === 1 && dst.owner === opp && dst.count === 1) {
        preview.bar[oppIdx] = (preview.bar[oppIdx] || 0) + 1;
        dst.owner = 0;
        dst.count = 0;
      }
      dst.owner = player;
      dst.count = (dst.count || 0) + 1;
    }

    return preview;
  }

  function currentBoardState() {
    if (!state.game) {
      return null;
    }
    if (state.click.path.length === 0) {
      return state.game;
    }
    try {
      return applyPreviewLine(state.game, state.click.path);
    } catch (_) {
      return state.game;
    }
  }

  function rectCenterRect(rect, size = 34) {
    if (!rect) {
      return null;
    }
    const side = Math.min(size, rect.width, rect.height);
    return {
      left: rect.left + rect.width / 2 - side / 2,
      top: rect.top + rect.height / 2 - side / 2,
      width: side,
      height: side,
    };
  }

  function moveAnchorRect(pointIdx, player, toOff = false) {
    if (pointIdx === 0) {
      const selector = player === 1 ? ".barStack.bottom" : ".barStack.top";
      const barStack = ui.board.querySelector(selector) || ui.board.querySelector(".barColumn");
      return rectCenterRect(barStack ? barStack.getBoundingClientRect() : null);
    }
    if (toOff) {
      return rectCenterRect(ui.offInfo.getBoundingClientRect());
    }
    const point = ui.board.querySelector(`.point[data-point="${pointIdx}"]`);
    if (!point) {
      return null;
    }
    return rectCenterRect(point.getBoundingClientRect());
  }

  function queueMoveAnimation(move, player = state.game ? state.game.turn : 0) {
    if (!move || !player) {
      return;
    }
    const colorClass = checkerColorClassForPlayer(player);
    const startRect = moveAnchorRect(move.from, player, false);
    if (!startRect || !colorClass) {
      return;
    }
    state.animation.queue.push({
      move,
      colorClass,
      player,
      startRect: {
        left: startRect.left,
        top: startRect.top,
        width: startRect.width,
        height: startRect.height,
      },
    });
  }

  function queueLineAnimation(moves, player) {
    if (!Array.isArray(moves) || moves.length === 0) {
      return;
    }
    moves.forEach((move) => queueMoveAnimation(move, player));
  }

  function playPendingMoveAnimation() {
    if (state.animation.playing || state.animation.queue.length === 0) {
      return;
    }
    const pending = state.animation.queue.shift();
    state.animation.playing = true;
    const endRect = moveAnchorRect(pending.move.to, pending.player, pending.move.to === 0);
    if (!endRect) {
      state.animation.playing = false;
      if (state.animation.queue.length > 0) {
        window.setTimeout(playPendingMoveAnimation, 40);
      }
      return;
    }

    const ghost = document.createElement("div");
    ghost.className = `checker checkerGhost ${pending.colorClass}`;
    ghost.style.left = `${pending.startRect.left}px`;
    ghost.style.top = `${pending.startRect.top}px`;
    ghost.style.width = `${pending.startRect.width}px`;
    ghost.style.height = `${pending.startRect.height}px`;
    document.body.appendChild(ghost);

    const dx = endRect.left - pending.startRect.left;
    const dy = endRect.top - pending.startRect.top;
    requestAnimationFrame(() => {
      ghost.style.transform = `translate(${dx}px, ${dy}px)`;
      ghost.style.opacity = "0.2";
    });
    window.setTimeout(() => {
      ghost.remove();
      state.animation.playing = false;
      if (state.animation.queue.length > 0) {
        window.setTimeout(playPendingMoveAnimation, 90);
      }
    }, 280);
  }

  function createPoint(boardState, pointIdx, rowType, colIndex) {
    const pointState = boardState.points[pointIdx] || { owner: 0, count: 0 };

    const point = document.createElement("div");
    const tone = colIndex % 2 === 0 ? "tone-a" : "tone-b";
    point.className = `point ${rowType} ${tone}`;
    point.style.gridColumn = String(colIndex + 1);
    point.style.gridRow = rowType === "top" ? "1" : "2";
    point.dataset.point = String(pointIdx);

    if (state.edit.enabled) {
      point.classList.add("interactive");
      if (canSelectEditSource(pointIdx)) {
        point.classList.add("fromOption");
      }
      if (canSelectEditDestination(pointIdx)) {
        point.classList.add("toOption");
      }
      if (state.edit.source && state.edit.source.from === pointIdx) {
        point.classList.add("selectedFrom");
      }
    } else if (!state.isBotTurn && state.legalLines.length > 0) {
      point.classList.add("interactive");
      if (state.click.selectableFrom.has(pointIdx)) {
        point.classList.add("fromOption");
      }
      if (state.click.selectableTo.has(pointIdx)) {
        point.classList.add("toOption");
      }
      if (state.click.selectedFrom === pointIdx) {
        point.classList.add("selectedFrom");
      }
    }

    if (pointIsInPath(pointIdx)) {
      point.classList.add("inPath");
    }

    if (pointState.owner !== 0) {
      point.classList.add(`owner-${ownerClass(pointState.owner)}`);
    }

    point.addEventListener("click", () => onPointClick(pointIdx));
    point.addEventListener("contextmenu", (event) => {
      event.preventDefault();
      onBoardSecondaryAction();
    });

    const tri = document.createElement("div");
    tri.className = "triangle";
    point.appendChild(tri);

    const stack = document.createElement("div");
    stack.className = "checkerStack";
    drawStack(stack, ownerClass(pointState.owner), pointState.count, rowType === "top" ? "top" : "bottom", 15);
    point.appendChild(stack);

    const label = document.createElement("span");
    label.className = "label";
    label.textContent = String(pointIdx);
    point.appendChild(label);

    return point;
  }

  function createBar(boardState) {
    const bar = document.createElement("div");
    bar.className = "barColumn";

    const blackCount = (boardState.bar && boardState.bar[1]) || 0;
    const whiteCount = (boardState.bar && boardState.bar[0]) || 0;

    const topLabel = document.createElement("div");
    topLabel.className = "barLabel top";
    topLabel.textContent = t("bar.black", { count: blackCount });
    bar.appendChild(topLabel);

    const topStack = document.createElement("div");
    topStack.className = "barStack top";
    drawStack(topStack, "black", blackCount, "top", 8);
    bar.appendChild(topStack);

    const bottomLabel = document.createElement("div");
    bottomLabel.className = "barLabel bottom";
    bottomLabel.textContent = t("bar.white", { count: whiteCount });
    bar.appendChild(bottomLabel);

    const bottomStack = document.createElement("div");
    bottomStack.className = "barStack bottom";
    drawStack(bottomStack, "white", whiteCount, "bottom", 8);
    bar.appendChild(bottomStack);

    if (state.edit.enabled) {
      if (canSelectEditDestination(-1)) {
        bar.classList.add("interactive", "toOption");
      }
      bar.addEventListener("click", () => onEditTargetClick(-1));
      bar.addEventListener("contextmenu", (event) => {
        event.preventDefault();
        onBoardSecondaryAction();
      });
      return bar;
    }

    if (!state.isBotTurn && state.legalLines.length > 0) {
      bar.classList.add("interactive");
      if (state.click.selectableFrom.has(0)) {
        bar.classList.add("fromOption");
      }
      if (state.click.selectedFrom === 0) {
        bar.classList.add("selectedFrom");
      }
    }
    bar.addEventListener("click", () => onPointClick(0));
    bar.addEventListener("contextmenu", (event) => {
      event.preventDefault();
      onBoardSecondaryAction();
    });

    return bar;
  }

  function refreshSnapshot() {
    if (!state.game) {
      ui.snapshot.textContent = t("common.none");
      return;
    }
    ui.snapshot.textContent = JSON.stringify(
      {
        game_type: state.game.game_type,
        turn: state.game.turn,
        move_number: state.game.meta ? state.game.meta.move_number : 0,
        off: state.game.off,
        bar: state.game.bar,
      },
      null,
      2
    );
  }

  function renderBoard() {
    ui.board.innerHTML = "";
    const boardState = currentBoardState();
    if (!boardState || !boardState.points) {
      return;
    }

    const bar = boardState.bar || [0, 0];
    const off = boardState.off || [0, 0];
    ui.barInfo.textContent = t("stats.bar", { white: bar[0] || 0, black: bar[1] || 0 });
    ui.offInfo.textContent = t("stats.off", { white: off[0] || 0, black: off[1] || 0 });
    refreshOffTarget();

    ui.board.appendChild(createBar(boardState));

    topRow.forEach((pointIdx, colIdx) => {
      if (pointIdx == null) {
        return;
      }
      ui.board.appendChild(createPoint(boardState, pointIdx, "top", colIdx));
    });

    bottomRow.forEach((pointIdx, colIdx) => {
      if (pointIdx == null) {
        return;
      }
      ui.board.appendChild(createPoint(boardState, pointIdx, "bottom", colIdx));
    });

    playPendingMoveAnimation();
  }

  function refreshOffTarget() {
    ui.offInfo.classList.remove("interactive", "toOption");
    ui.offInfo.onclick = null;
    ui.offInfo.oncontextmenu = (event) => {
      event.preventDefault();
      onBoardSecondaryAction();
    };
    if (state.edit.enabled) {
      if (canSelectEditDestination(0)) {
        ui.offInfo.classList.add("interactive", "toOption");
        ui.offInfo.onclick = () => onEditTargetClick(0);
      }
      return;
    }
    if (!state.isBotTurn && state.legalLines.length > 0 && state.click.selectableTo.has(0)) {
      ui.offInfo.classList.add("interactive", "toOption");
      ui.offInfo.onclick = () => onPointClick(0);
    }
  }

  function renderTop3(decision) {
    ui.top3.innerHTML = "";
    if (!decision || !Array.isArray(decision.top3) || !decision.top3.length || !ui.showTop3.checked) {
      const li = document.createElement("li");
      li.textContent = t("common.noTop3");
      ui.top3.appendChild(li);
      return;
    }
    decision.top3.forEach((ev, i) => {
      const li = document.createElement("li");
      li.textContent = `#${i + 1} ${formatLine(ev.line)} | WinProb=${(ev.winprob || 0).toFixed(3)} | sims=${ev.sims || 0}`;
      ui.top3.appendChild(li);
    });
  }

  function translateCategory(category) {
    if (!category) {
      return "";
    }
    return t(`category.${category}`);
  }

  function renderAnalysis(data) {
    if (!data) {
      ui.analysis.textContent = t("common.noAnalysis");
      return;
    }
    ui.analysis.textContent = `${t("analysis.category")}: ${translateCategory(data.category)}\n${t("analysis.delta")}: ${Number(data.delta || 0).toFixed(4)}\n${t("analysis.best")}: ${formatLine(data.best_line)}\n${t("analysis.bestProb")}: ${Number(data.best_winprob || 0).toFixed(3)}`;
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
      op.textContent = t("common.noMoves");
      ui.lineIndex.appendChild(op);
      return;
    }

    state.legalLines.forEach((line, i) => {
      const op = document.createElement("option");
      op.value = String(i);
      op.textContent = t("line.option", { index: i + 1, line: formatLine(line) });
      ui.lineIndex.appendChild(op);
    });

    if (state.legalLines.length === 1) {
      ui.lineIndex.value = "0";
    }
  }

  function syncTurnStatus(resp) {
    const turn = resp && resp.state ? resp.state.turn : 0;
    const side = turn === 1 ? t("side.white") : turn === 2 ? t("side.black") : "?";
    const botTurn = !!(resp && resp.isBotTurn);
    state.isBotTurn = botTurn;
    setStatus(t("status.turn", { side, actor: botTurn ? t("actor.bot") : t("actor.human") }));
    updateTurnGuide();
  }

  function updateTurnGuide() {
    if (!ui.turnGuide) {
      return;
    }
    if (!state.game) {
      ui.turnGuide.textContent = t("guide.idle");
      return;
    }
    if (state.busy) {
      ui.turnGuide.textContent = t("status.busy");
      return;
    }
    if (state.edit.enabled) {
      ui.turnGuide.textContent = t("guide.edit");
      return;
    }
    if (state.isBotTurn) {
      ui.turnGuide.textContent = botAutoPlayEnabled() ? t("guide.bot") : t("guide.botManual");
      return;
    }
    if (state.legalLines.length > 0) {
      ui.turnGuide.textContent = t("guide.humanMove");
      return;
    }
    ui.turnGuide.textContent = t("guide.humanDice");
  }

  function setBusy(nextBusy) {
    state.busy = nextBusy;
    const disabled = !!nextBusy;
    ui.startBtn.disabled = disabled;
    ui.swapSideBtn.disabled = disabled;
    ui.applyBtn.disabled = disabled;
    ui.legalBtn.disabled = disabled;
    ui.clearSelectionBtn.disabled = disabled;
    if (ui.editBtn) {
      ui.editBtn.disabled = disabled;
    }
    ui.undoBtn.disabled = disabled;
    ui.exportBtn.disabled = disabled;
    if (ui.selfLearnBtn) {
      ui.selfLearnBtn.disabled = disabled;
    }
    if (ui.bgTrainingBtn) {
      ui.bgTrainingBtn.disabled = disabled;
    }
    if (ui.syncDiceBtn) {
      ui.syncDiceBtn.disabled = disabled;
    }
    renderDiceTray();
    updateTurnGuide();
  }

  function resetClickSelection(keepDice = true) {
    if (!keepDice) {
      state.click.d1 = 0;
      state.click.d2 = 0;
    }
    state.click.path = [];
    state.click.selectedFrom = null;
    state.click.candidates = state.legalLines.map((_, i) => i);
    state.click.selectableFrom = new Set();
    state.click.selectableTo = new Set();
    recomputeSelectable();
    updateSelectedPath();
    updateTurnGuide();
  }

  function refreshAutoPlayAvailability() {
    if (!ui.autoHumanRoll || !ui.opponent) {
      return;
    }
    const enabled = ui.opponent.value === "bot";
    ui.autoHumanRoll.disabled = !enabled;
    if (!enabled) {
      ui.autoHumanRoll.checked = false;
    }
  }

  function autoDiceEnabled() {
    return !!(ui.autoHumanRoll && ui.autoHumanRoll.checked);
  }

  function botAutoPlayEnabled() {
    return autoDiceEnabled() && ui.opponent && ui.opponent.value === "bot";
  }

  function isTerminalGame(game) {
    if (!game || !Array.isArray(game.off)) {
      return false;
    }
    return Number(game.off[0] || 0) >= 15 || Number(game.off[1] || 0) >= 15;
  }

  function scheduleAutoDiceNext(delay = 320) {
    if (!botAutoPlayEnabled() || !state.game || isTerminalGame(state.game)) {
      return;
    }
    window.setTimeout(() => {
      if (!state.busy && state.game && botAutoPlayEnabled() && !isTerminalGame(state.game)) {
        randomDice();
      }
    }, delay);
  }

  function prepareInteractiveLines(lines, d1, d2) {
    renderLegalLines(lines);
    state.click.d1 = d1;
    state.click.d2 = d2;
    resetClickSelection(true);
    renderBoard();
    updateTurnGuide();
  }

  function isSelectionComplete() {
    if (state.click.candidates.length !== 1) {
      return false;
    }
    const line = state.legalLines[state.click.candidates[0]];
    if (!line || !Array.isArray(line.moves)) {
      return false;
    }
    return state.click.path.length === line.moves.length;
  }

  function recomputeSelectable() {
    state.click.selectableFrom = new Set();
    state.click.selectableTo = new Set();

    if (!state.game || state.isBotTurn || state.legalLines.length === 0) {
      return;
    }

    const candidates = state.click.candidates.length ? state.click.candidates : state.legalLines.map((_, i) => i);
    const step = state.click.path.length;

    if (!candidates.length) {
      return;
    }

    for (const idx of candidates) {
      const line = state.legalLines[idx];
      const move = line && Array.isArray(line.moves) ? line.moves[step] : null;
      if (!move) {
        continue;
      }
      if (state.click.selectedFrom == null) {
        state.click.selectableFrom.add(move.from);
      } else if (move.from === state.click.selectedFrom) {
        state.click.selectableTo.add(move.to);
      }
    }
  }

  function rebuildCandidatesFromPath() {
    const prefix = state.click.path;
    state.click.candidates = state.legalLines
      .map((_, idx) => idx)
      .filter((idx) => {
        const line = state.legalLines[idx];
        if (!line || !Array.isArray(line.moves) || line.moves.length < prefix.length) {
          return false;
        }
        for (let i = 0; i < prefix.length; i++) {
          const want = prefix[i];
          const got = line.moves[i];
          if (!got || got.from !== want.from || got.to !== want.to || got.die !== want.die) {
            return false;
          }
        }
        return true;
      });
    if (state.click.candidates.length === 0) {
      state.click.candidates = state.legalLines.map((_, i) => i);
      state.click.path = [];
    }
  }

  function popLastSelectedStep() {
    if (state.click.path.length === 0) {
      return false;
    }
    state.click.path.pop();
    state.click.selectedFrom = null;
    rebuildCandidatesFromPath();
    recomputeSelectable();
    updateSelectedPath();
    renderBoard();
    setStatus(t("status.stepRemoved"));
    return true;
  }

  function updateSelectedPath() {
    if (state.edit.enabled) {
      if (!state.edit.source) {
        ui.selectedPath.textContent = t("path.editIdle");
      } else {
        ui.selectedPath.textContent = t("path.editSource", { from: formatPoint(state.edit.source.from) });
      }
      renderDiceTray();
      return;
    }
    if (state.click.path.length === 0) {
      ui.selectedPath.textContent = t("path.empty");
      renderDiceTray();
      return;
    }
    let msg = t("path.current", { path: state.click.path.map((mv) => formatMove(mv)).join(" ") });
    if (isSelectionComplete()) {
      msg += ` ${t("path.ready")}`;
    } else if (state.click.candidates.length > 1) {
      msg += ` ${t("path.variants", { count: state.click.candidates.length })}`;
    }
    ui.selectedPath.textContent = msg;
    renderDiceTray();
  }

  function renderEditModeState() {
    if (!ui.editBtn) {
      return;
    }
    ui.editBtn.classList.toggle("active", !!state.edit.enabled);
  }

  function resetEditSelection() {
    state.edit.source = null;
    updateSelectedPath();
  }

  function exitEditMode() {
    if (!state.edit.enabled && !state.edit.source) {
      return;
    }
    state.edit.enabled = false;
    resetEditSelection();
    renderEditModeState();
  }

  function requireEditModeOff() {
    if (!state.edit.enabled) {
      return false;
    }
    setStatus(t("status.editExitFirst"), true);
    return true;
  }

  function canSelectEditSource(pointIdx) {
    if (!state.edit.enabled || !state.game || pointIdx < 1 || pointIdx > 24) {
      return false;
    }
    const point = state.game.points[pointIdx];
    return !!(point && point.count > 0 && point.owner !== 0);
  }

  function canSelectEditDestination(pointIdx) {
    if (!state.edit.enabled || !state.edit.source || !state.game) {
      return false;
    }
    if (pointIdx === state.edit.source.from) {
      return false;
    }
    if (pointIdx === 0) {
      return true;
    }
    if (pointIdx === -1) {
      return state.game.game_type === 1;
    }
    if (pointIdx < 1 || pointIdx > 24) {
      return false;
    }
    const point = state.game.points[pointIdx] || { owner: 0, count: 0 };
    return point.owner === 0 || point.owner === state.edit.source.color;
  }

  async function applyEditChecker(targetIdx) {
    const source = state.edit.source;
    if (!source) {
      setStatus(t("status.editNeedChecker"), true);
      return;
    }
    setBusy(true);
    try {
      const resp = await call("EditChecker", { from: source.from, to: targetIdx, color: source.color === 1 ? "white" : "black" });
      state.game = resp.state;
      state.isBotTurn = !!resp.isBotTurn;
      renderLegalLines([]);
      resetClickSelection(true);
      resetEditSelection();
      renderBoard();
      refreshSnapshot();
      renderTop3(null);
      renderAnalysis(null);
      renderEditModeState();
      setStatus(t("status.editDone", { from: formatPoint(source.from), to: formatPoint(targetIdx) }));
      pushLog(t("log.editChecker", { from: formatPoint(source.from), to: formatPoint(targetIdx) }));
    } catch (err) {
      setStatus(translateErrorMessage(String(err)), true);
    } finally {
      setBusy(false);
      updateTurnGuide();
    }
  }

  async function onEditTargetClick(targetIdx) {
    if (!state.game) {
      setStatus(t("status.noGame"), true);
      return;
    }
    if (!state.edit.source) {
      setStatus(t("status.editNeedChecker"), true);
      return;
    }
    if (!canSelectEditDestination(targetIdx)) {
      setStatus(t("status.editBadDestination"), true);
      return;
    }
    await applyEditChecker(targetIdx);
  }

  async function toggleEditMode() {
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    if (!state.game) {
      setStatus(t("status.noGame"), true);
      return;
    }
    state.edit.enabled = !state.edit.enabled;
    resetEditSelection();
    renderLegalLines([]);
    resetClickSelection(true);
    renderTop3(null);
    renderAnalysis(null);
    renderBoard();
    renderEditModeState();
    setStatus(t(state.edit.enabled ? "status.editModeOn" : "status.editCancelled"));
    updateTurnGuide();
  }


  async function onBoardSecondaryAction() {
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    if (state.edit.enabled) {
      if (state.edit.source) {
        resetEditSelection();
        renderBoard();
        setStatus(t("status.editModeOn"));
      }
      return;
    }
    if (popLastSelectedStep()) {
      return;
    }
    await undo();
  }

  async function onPointClick(pointIdx) {
    if (state.edit.enabled) {
      if (!state.game) {
        setStatus(t("status.noGame"), true);
        return;
      }
      if (state.busy) {
        setStatus(t("status.busy"));
        return;
      }
      if (pointIdx < 1 || pointIdx > 24) {
        setStatus(t("status.editSourceOnly"), true);
        return;
      }
      if (!state.edit.source) {
        if (!canSelectEditSource(pointIdx)) {
          setStatus(t("status.editNeedChecker"), true);
          return;
        }
        const point = state.game.points[pointIdx];
        state.edit.source = { from: pointIdx, color: point.owner };
        updateSelectedPath();
        renderBoard();
        setStatus(t("status.editPickDestination", { point: formatPoint(pointIdx) }));
        updateTurnGuide();
        return;
      }
      if (pointIdx === state.edit.source.from) {
        resetEditSelection();
        renderBoard();
        setStatus(t("status.editModeOn"));
        updateTurnGuide();
        return;
      }
      if (!canSelectEditDestination(pointIdx)) {
        setStatus(t("status.editBadDestination"), true);
        return;
      }
      await applyEditChecker(pointIdx);
      return;
    }

    if (!state.game) {
      setStatus(t("status.noGame"), true);
      return;
    }
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    if (state.isBotTurn) {
      setStatus(t("status.waitBot"), true);
      return;
    }
    if (state.legalLines.length === 0) {
      setStatus(t("status.loadLegalFirst"), true);
      return;
    }
    if (isSelectionComplete()) {
      setStatus(t("status.selectionLocked"));
      return;
    }

    if (state.click.selectedFrom == null) {
      if (!state.click.selectableFrom.has(pointIdx)) {
        setStatus(t("status.pickHighlighted"), true);
        return;
      }
      state.click.selectedFrom = pointIdx;
      recomputeSelectable();
      renderBoard();
      setStatus(t("status.pickDestination", { point: formatPoint(pointIdx) }));
      return;
    }

    if (pointIdx === state.click.selectedFrom) {
      state.click.selectedFrom = null;
      recomputeSelectable();
      renderBoard();
      setStatus(t("status.pickSource"));
      return;
    }

    if (!state.click.selectableTo.has(pointIdx)) {
      setStatus(t("status.badDestination"), true);
      return;
    }

    const step = state.click.path.length;
    const narrowed = state.click.candidates.filter((idx) => {
      const move = state.legalLines[idx] && state.legalLines[idx].moves ? state.legalLines[idx].moves[step] : null;
      return !!move && move.from === state.click.selectedFrom && move.to === pointIdx;
    });

    if (!narrowed.length) {
      setStatus(t("status.badDestination"), true);
      return;
    }

    const chosenMove = state.legalLines[narrowed[0]].moves[step];
    queueMoveAnimation(chosenMove);
    state.click.path.push(chosenMove);
    state.click.candidates = narrowed;
    state.click.selectedFrom = null;

    if (isSelectionComplete()) {
      ui.lineIndex.value = String(state.click.candidates[0]);
      setStatus(t("status.lineSelected", { index: state.click.candidates[0] + 1 }));
    } else {
      setStatus(t("status.stepAccepted", { from: chosenMove.from, to: chosenMove.to, die: chosenMove.die }));
    }

    recomputeSelectable();
    updateSelectedPath();
    renderBoard();
    if (isSelectionComplete()) {
      await applyDice(true);
    }
  }

  async function ensureLegalLinesForDice(d1, d2) {
    if (state.click.d1 === d1 && state.click.d2 === d2 && state.legalLines.length > 0) {
      return state.legalLines;
    }
    const lines = await call("LegalLines", d1, d2);
    prepareInteractiveLines(lines, d1, d2);
    return lines;
  }

  function handleApplyResponse(resp, d1, d2) {
    exitEditMode();
    const prevState = state.game ? cloneGameStateForPreview(state.game) : null;
    if (prevState && resp && resp.decision && resp.decision.chosen_line && Array.isArray(resp.decision.chosen_line.moves)) {
      queueLineAnimation(resp.decision.chosen_line.moves, prevState.turn);
    }
    if (prevState && state.click.path.length === 0 && resp && !resp.decision && resp.applied && Array.isArray(resp.applied.moves) && resp.applied.moves.length > 0) {
      queueLineAnimation(resp.applied.moves, prevState.turn);
    }

    state.game = resp.state;
    state.isBotTurn = !!resp.isBotTurn;

    renderLegalLines([]);
    resetClickSelection(false);
    renderBoard();
    refreshSnapshot();
    syncTurnStatus(resp);

    if (resp.decision) {
      renderTop3(resp.decision);
      renderAnalysis(null);
      pushLog(
        t("log.botMove", {
          d1,
          d2,
          line: formatLine(resp.decision.chosen_line),
          prob: Number(resp.decision.chosen_prob || 0).toFixed(3),
        })
      );
      setStatus(t("status.autoBotMove"));
      scheduleAutoDiceNext();
      return;
    }

    renderTop3(null);
    renderAnalysis(resp.analysis || null);
    if (resp.applied) {
      pushLog(t("log.humanMove", { d1, d2, line: formatLine(resp.applied) }));
    } else {
      pushLog(t("log.pass", { d1, d2 }));
      setStatus(t("status.autoNoMoves"));
    }

    scheduleAutoDiceNext();
  }

  async function startGame() {
    if (state.busy) {
      return;
    }
    setBusy(true);
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
      exitEditMode();
      state.game = resp.state;
      state.isBotTurn = !!resp.isBotTurn;
      renderLegalLines([]);
      resetClickSelection(false);
      renderTop3(null);
      renderAnalysis(null);
      await setDiceValues(0, 0, false);
      refreshSnapshot();
      renderBoard();
      syncTurnStatus(resp);
      pushLog(
        t("log.newGame", {
          game: req.gameType,
          opponent: req.opponent,
          bot: req.botSide,
          think: req.thinkTime,
        })
      );
      if (botAutoPlayEnabled()) {
        scheduleAutoDiceNext();
      }
    } catch (err) {
      setStatus(translateErrorMessage(String(err)), true);
    } finally {
      setBusy(false);
      updateTurnGuide();
    }
  }

  async function loadLegal() {
    if (requireEditModeOff()) {
      return;
    }
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    setBusy(true);
    try {
      if (!state.game) {
        setStatus(t("status.noGame"), true);
        return;
      }
      const d1 = safeInt(ui.d1.value, 0);
      const d2 = safeInt(ui.d2.value, 0);
      if (d1 < 1 || d1 > 6 || d2 < 1 || d2 > 6) {
        setStatus(t("errors.dice"), true);
        return;
      }
      const lines = await call("LegalLines", d1, d2);
      prepareInteractiveLines(lines, d1, d2);
      if (!state.isBotTurn && lines.length > 0) {
        setStatus(t("status.autoHumanReady"));
      } else if (lines.length === 0) {
        setStatus(t("status.noLegal"));
      } else {
        setStatus(t("status.legalCount", { count: lines.length }));
      }
    } catch (err) {
      setStatus(translateErrorMessage(String(err)), true);
    } finally {
      setBusy(false);
      updateTurnGuide();
    }
  }

  async function applyDice(autoTriggered = false) {
    if (requireEditModeOff()) {
      return;
    }
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    setBusy(true);
    try {
      if (!state.game) {
        setStatus(t("status.noGame"), true);
        return;
      }

      const d1 = safeInt(ui.d1.value, 0);
      const d2 = safeInt(ui.d2.value, 0);
      if (d1 < 1 || d1 > 6 || d2 < 1 || d2 > 6) {
        setStatus(t("errors.dice"), true);
        return;
      }

      const current = await call("State");
      state.isBotTurn = !!current.isBotTurn;
      let lineIndex = -1;

      if (!state.isBotTurn) {
        const lines = await ensureLegalLinesForDice(d1, d2);

        if (lines.length > 1) {
          if (isSelectionComplete()) {
            lineIndex = state.click.candidates[0];
          } else {
            const fromSelect = safeInt(ui.lineIndex.value, -1);
            if (fromSelect >= 0 && fromSelect < lines.length) {
              lineIndex = fromSelect;
            }
          }
          if (lineIndex < 0 || lineIndex >= lines.length) {
            setStatus(t("status.chooseLineAgain"), true);
            return;
          }
        }
        if (lines.length === 1) {
          lineIndex = 0;
          ui.lineIndex.value = "0";
        }
      }

      const resp = await call("ApplyDice", d1, d2, lineIndex);
      handleApplyResponse(resp, d1, d2);
      if (autoTriggered && !resp.decision && resp.applied) {
        setStatus(t("status.autoApplied"));
      }
    } catch (err) {
      setStatus(translateErrorMessage(String(err)), true);
    } finally {
      setBusy(false);
      updateTurnGuide();
    }
  }

  async function undo() {
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    setBusy(true);
    try {
      const resp = await call("Undo");
      state.game = resp.state;
      state.isBotTurn = !!resp.isBotTurn;
      renderLegalLines([]);
      resetClickSelection(false);
      renderBoard();
      refreshSnapshot();
      renderTop3(null);
      renderAnalysis(null);
      syncTurnStatus(resp);
      pushLog(t("log.undo"));
    } catch (err) {
      setStatus(translateErrorMessage(String(err)), true);
    } finally {
      setBusy(false);
      updateTurnGuide();
    }
  }

  async function swapBotSide() {
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    setBusy(true);
    try {
      if (requireEditModeOff()) {
        return;
      }
      const resp = await call("SwapBotSide");
      state.game = resp.state;
      state.isBotTurn = !!resp.isBotTurn;
      renderLegalLines([]);
      resetClickSelection(false);
      renderBoard();
      refreshSnapshot();
      renderTop3(null);
      renderAnalysis(null);
      syncTurnStatus(resp);
      const botSide = ui.botSide.value === "white" ? "black" : "white";
      ui.botSide.value = botSide;
      const sideLabel = botSide === "white" ? t("side.white") : t("side.black");
      setStatus(t("status.sideSwapped", { side: sideLabel }));
      pushLog(t("log.swapSide", { side: sideLabel }));
    } catch (err) {
      setStatus(translateErrorMessage(String(err)), true);
    } finally {
      setBusy(false);
      updateTurnGuide();
    }
  }

  async function exportState() {
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    try {
      const path = prompt(t("prompt.export"), "state.json");
      if (!path) {
        return;
      }
      await call("Export", path);
      setStatus(t("status.exported", { path }));
      pushLog(t("log.export", { path }));
    } catch (err) {
      setStatus(translateErrorMessage(String(err)), true);
    }
  }

  async function randomDice() {
    if (requireEditModeOff()) {
      return;
    }
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    const d1 = Math.floor(Math.random() * 6) + 1;
    const d2 = Math.floor(Math.random() * 6) + 1;
    await setDiceValues(d1, d2, false);

    if (!state.game) {
      setStatus(t("status.randomReady", { d1, d2 }));
      return;
    }

    setBusy(true);
    try {
      renderLegalLines([]);
      resetClickSelection(false);
      renderBoard();
      setStatus(t("status.randomReady", { d1, d2 }));

      const current = await call("State");
      state.isBotTurn = !!current.isBotTurn;

      if (state.isBotTurn) {
        const resp = await call("ApplyDice", d1, d2, -1);
        handleApplyResponse(resp, d1, d2);
        return;
      }

      const lines = await call("LegalLines", d1, d2);
      prepareInteractiveLines(lines, d1, d2);
      if (lines.length === 0) {
        const resp = await call("ApplyDice", d1, d2, -1);
        handleApplyResponse(resp, d1, d2);
      } else if (lines.length === 1) {
        ui.lineIndex.value = "0";
        state.click.candidates = [0];
        state.click.path = Array.isArray(lines[0].moves) ? lines[0].moves.slice() : [];
        recomputeSelectable();
        updateSelectedPath();
        renderBoard();
        const resp = await call("ApplyDice", d1, d2, 0);
        handleApplyResponse(resp, d1, d2);
        setStatus(t("status.autoApplied"));
      } else {
        setStatus(t("status.autoHumanReady"));
      }
    } catch (err) {
      setStatus(translateErrorMessage(String(err)), true);
    } finally {
      if (state.busy) {
        setBusy(false);
      }
      updateTurnGuide();
    }
  }

  async function applyDiceText() {
    if (requireEditModeOff()) {
      return;
    }
    const parsed = parseDiceText(ui.diceInput.value);
    if (!parsed) {
      setStatus(t("errors.dice"), true);
      return;
    }
    const isBotTurnNow = !!state.isBotTurn;
    await setDiceValues(parsed.d1, parsed.d2, !isBotTurnNow);
    if (isBotTurnNow) {
      await applyDice();
    }
  }

  async function onDieButton() {
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    await randomDice();
  }

  function clearSelection() {
    if (state.busy) {
      setStatus(t("status.busy"));
      return;
    }
    if (state.edit.enabled) {
      resetEditSelection();
      renderBoard();
      setStatus(t("status.editModeOn"));
      updateTurnGuide();
      return;
    }
    resetClickSelection(true);
    renderBoard();
    setStatus(t("status.selectionCleared"));
    updateTurnGuide();
  }

  function onLineSelected() {
    if (state.busy) {
      return;
    }
    if (requireEditModeOff()) {
      return;
    }
    const idx = safeInt(ui.lineIndex.value, -1);
    if (idx < 0 || idx >= state.legalLines.length) {
      resetClickSelection(true);
      renderBoard();
      updateTurnGuide();
      return;
    }
    const line = state.legalLines[idx];
    state.click.candidates = [idx];
    state.click.path = Array.isArray(line.moves) ? line.moves.slice() : [];
    state.click.selectedFrom = null;
    recomputeSelectable();
    updateSelectedPath();
    renderBoard();
    updateTurnGuide();
  }

  async function onDiceChanged() {
    if (state.edit.enabled) {
      syncDiceInputText();
      renderDiceTray();
      return;
    }
    if (!state.game) {
      return;
    }
    if (state.busy) {
      return;
    }
    syncDiceInputText();
    renderDiceTray();
    const d1 = safeInt(ui.d1.value, 0);
    const d2 = safeInt(ui.d2.value, 0);
    if (d1 === state.click.d1 && d2 === state.click.d2) {
      return;
    }
    renderLegalLines([]);
    resetClickSelection(false);
    renderBoard();
    if (d1 < 1 || d1 > 6 || d2 < 1 || d2 > 6) {
      setStatus(t("status.diceChanged"));
      updateTurnGuide();
      return;
    }
    if (state.isBotTurn) {
      setStatus(t("status.diceChanged"));
      updateTurnGuide();
      return;
    }
    const reqSeq = ++state.diceRequestSeq;
    try {
      const lines = await call("LegalLines", d1, d2);
      if (reqSeq !== state.diceRequestSeq) {
        return;
      }
      prepareInteractiveLines(lines, d1, d2);
      if (lines.length > 0) {
        setStatus(t("status.autoHumanReady"));
      } else {
        setStatus(t("status.noLegal"));
      }
    } catch (err) {
      if (reqSeq !== state.diceRequestSeq) {
        return;
      }
      setStatus(translateErrorMessage(String(err)), true);
    }
    updateTurnGuide();
  }

  function setSelectText(select, mapping) {
    Array.from(select.options).forEach((option) => {
      if (mapping[option.value]) {
        option.textContent = mapping[option.value];
      }
    });
  }

  function applyLocale() {
    document.documentElement.lang = state.lang;
    document.title = t("hero.title");

    document.querySelectorAll("[data-i18n]").forEach((el) => {
      const key = el.dataset.i18n;
      el.textContent = t(key);
    });

    ui.seed.placeholder = t("labels.seedPlaceholder");

    setSelectText(ui.gameType, { short: t("options.short"), long: t("options.long") });
    setSelectText(ui.botSide, { white: t("options.white"), black: t("options.black") });
    setSelectText(ui.opponent, { human: t("options.human"), bot: t("options.bot") });
    refreshAutoPlayAvailability();

    renderLegalLines(state.legalLines);
    updateSelectedPath();
    renderBoard();
    syncDiceInputText();
    renderDiceTray();
    renderAnalysis(null);
    renderTrainingStatus();
    renderEditModeState();

    if (!state.game) {
      setStatus(t("status.ready"));
      updateTurnGuide();
    } else {
      syncTurnStatus({ state: state.game, isBotTurn: state.isBotTurn });
    }
  }

  function translateErrorMessage(raw) {
    const msg = String(raw || "");
    const m = msg.toLowerCase();

    if (m.includes("wails api")) return t("errors.wails");
    if (m.includes("dice must be in 1..6") || m.includes("invalid dice")) return t("errors.dice");
    if (m.includes("line index required")) return t("errors.lineRequired");
    if (m.includes("current turn belongs to bot")) return t("errors.turnBot");
    if (m.includes("current turn belongs to human")) return t("errors.turnHuman");
    if (m.includes("nothing to undo")) return t("errors.undo");
    if (m.includes("export path is empty")) return t("errors.exportPath");
    if (m.includes("line index out of range")) return t("errors.outOfRange");
    if (m.includes("illegal line for analysis")) return t("errors.illegalAnalysis");
    if (m.includes("edit point is out of range")) return t("errors.editRange");
    if (m.includes("source and destination are the same")) return t("errors.editSame");
    if (m.includes("source point is empty") || m.includes("borne-off source is empty") || m.includes("bar source is empty")) return t("errors.editEmpty");
    if (m.includes("destination point belongs to the opponent")) return t("errors.editOpponent");
    if (m.includes("no training examples found")) return t("errors.noTrainingData");
    if (m.includes("training") || m.includes("selflearn") || m.includes("self-learning")) return t("errors.training", { message: msg });
    return t("errors.generic", { message: msg });
  }

  ui.startBtn.addEventListener("click", startGame);
  ui.swapSideBtn.addEventListener("click", swapBotSide);
  ui.applyBtn.addEventListener("click", applyDice);
  ui.legalBtn.addEventListener("click", loadLegal);
  ui.clearSelectionBtn.addEventListener("click", clearSelection);
  if (ui.editBtn) {
    ui.editBtn.addEventListener("click", toggleEditMode);
  }
  ui.undoBtn.addEventListener("click", undo);
  ui.exportBtn.addEventListener("click", exportState);
  if (ui.selfLearnBtn) {
    ui.selfLearnBtn.addEventListener("click", selfLearn);
  }
  if (ui.bgTrainingBtn) {
    ui.bgTrainingBtn.addEventListener("click", toggleBackgroundTraining);
  }
  ui.lineIndex.addEventListener("change", onLineSelected);
  ui.syncDiceBtn.addEventListener("click", applyDiceText);
  ui.diceInput.addEventListener("change", applyDiceText);
  [ui.dieBtn1, ui.dieBtn2, ui.dieBtn3, ui.dieBtn4].forEach((btn) => btn.addEventListener("click", onDieButton));
  ui.opponent.addEventListener("change", () => {
    refreshAutoPlayAvailability();
  });
  ui.autoHumanRoll.addEventListener("change", () => {
    if (botAutoPlayEnabled() && state.game && !state.busy && !isTerminalGame(state.game)) {
      scheduleAutoDiceNext(80);
    }
    updateTurnGuide();
  });
  ui.langSelect.addEventListener("change", () => {
    state.lang = ui.langSelect.value === "en" ? "en" : "ru";
    applyLocale();
  });

  ui.langSelect.value = state.lang;
  applyLocale();
  renderTrainingStatus();
  refreshBackgroundTrainingStatus();
})();
