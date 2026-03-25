export namespace app {
	
	export class AnalysisResult {
	    category: string;
	    delta: number;
	    best_line: engine.TurnLine;
	    best_winprob: number;
	
	    static createFrom(source: any = {}) {
	        return new AnalysisResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.delta = source["delta"];
	        this.best_line = this.convertValues(source["best_line"], engine.TurnLine);
	        this.best_winprob = source["best_winprob"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BackgroundTrainingStatus {
	    running: boolean;
	    games: number;
	    examples: number;
	    workers: number;
	    // Go type: time
	    started_at: any;
	    last_train?: training.TrainSummary;
	    last_error?: string;
	    game_type: string;
	    think_time_sec: number;
	
	    static createFrom(source: any = {}) {
	        return new BackgroundTrainingStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.games = source["games"];
	        this.examples = source["examples"];
	        this.workers = source["workers"];
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.last_train = this.convertValues(source["last_train"], training.TrainSummary);
	        this.last_error = source["last_error"];
	        this.game_type = source["game_type"];
	        this.think_time_sec = source["think_time_sec"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ImportLogResult {
	    path: string;
	    imported: number;
	
	    static createFrom(source: any = {}) {
	        return new ImportLogResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.imported = source["imported"];
	    }
	}

}

export namespace bot {
	
	export class MoveEvaluation {
	    line: engine.TurnLine;
	    winprob: number;
	    sims: number;
	
	    static createFrom(source: any = {}) {
	        return new MoveEvaluation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.line = this.convertValues(source["line"], engine.TurnLine);
	        this.winprob = source["winprob"];
	        this.sims = source["sims"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Decision {
	    legal_count: number;
	    chosen_line: engine.TurnLine;
	    chosen_prob: number;
	    top3: MoveEvaluation[];
	    think_elapsed: number;
	    seed: number;
	
	    static createFrom(source: any = {}) {
	        return new Decision(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.legal_count = source["legal_count"];
	        this.chosen_line = this.convertValues(source["chosen_line"], engine.TurnLine);
	        this.chosen_prob = source["chosen_prob"];
	        this.top3 = this.convertValues(source["top3"], MoveEvaluation);
	        this.think_elapsed = source["think_elapsed"];
	        this.seed = source["seed"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace desktop {
	
	export class BackgroundTrainingResponse {
	    status: app.BackgroundTrainingStatus;
	
	    static createFrom(source: any = {}) {
	        return new BackgroundTrainingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = this.convertValues(source["status"], app.BackgroundTrainingStatus);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EditCheckerRequest {
	    from: number;
	    to: number;
	    color?: string;
	
	    static createFrom(source: any = {}) {
	        return new EditCheckerRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.to = source["to"];
	        this.color = source["color"];
	    }
	}
	export class ImportLogResponse {
	    result: app.ImportLogResult;
	
	    static createFrom(source: any = {}) {
	        return new ImportLogResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.result = this.convertValues(source["result"], app.ImportLogResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SelfLearnResponse {
	    result: training.TrainSummary;
	
	    static createFrom(source: any = {}) {
	        return new SelfLearnResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.result = this.convertValues(source["result"], training.TrainSummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StartRequest {
	    gameType: string;
	    botSide: string;
	    opponent: string;
	    thinkTime: number;
	    showTop3: boolean;
	    seed: number;
	    logPath: string;
	
	    static createFrom(source: any = {}) {
	        return new StartRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.gameType = source["gameType"];
	        this.botSide = source["botSide"];
	        this.opponent = source["opponent"];
	        this.thinkTime = source["thinkTime"];
	        this.showTop3 = source["showTop3"];
	        this.seed = source["seed"];
	        this.logPath = source["logPath"];
	    }
	}
	export class TurnResponse {
	    state: engine.GameState;
	    isBotTurn: boolean;
	    decision?: bot.Decision;
	    analysis?: app.AnalysisResult;
	    applied?: engine.TurnLine;
	
	    static createFrom(source: any = {}) {
	        return new TurnResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = this.convertValues(source["state"], engine.GameState);
	        this.isBotTurn = source["isBotTurn"];
	        this.decision = this.convertValues(source["decision"], bot.Decision);
	        this.analysis = this.convertValues(source["analysis"], app.AnalysisResult);
	        this.applied = this.convertValues(source["applied"], engine.TurnLine);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace engine {
	
	export class StateMeta {
	    move_number: number;
	
	    static createFrom(source: any = {}) {
	        return new StateMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.move_number = source["move_number"];
	    }
	}
	export class Point {
	    owner: number;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new Point(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.owner = source["owner"];
	        this.count = source["count"];
	    }
	}
	export class GameState {
	    game_type: number;
	    points: Point[];
	    off: number[];
	    bar: number[];
	    turn: number;
	    seed?: number;
	    meta: StateMeta;
	
	    static createFrom(source: any = {}) {
	        return new GameState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.game_type = source["game_type"];
	        this.points = this.convertValues(source["points"], Point);
	        this.off = source["off"];
	        this.bar = source["bar"];
	        this.turn = source["turn"];
	        this.seed = source["seed"];
	        this.meta = this.convertValues(source["meta"], StateMeta);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Move {
	    from: number;
	    to: number;
	    die: number;
	
	    static createFrom(source: any = {}) {
	        return new Move(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.to = source["to"];
	        this.die = source["die"];
	    }
	}
	
	
	export class TurnLine {
	    moves: Move[];
	
	    static createFrom(source: any = {}) {
	        return new TurnLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.moves = this.convertValues(source["moves"], Move);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace training {
	
	export class TrainSummary {
	    examples: number;
	    train_examples: number;
	    validation_examples: number;
	    epochs: number;
	    learning_rate: number;
	    l2: number;
	    avg_abs_error: number;
	    validation_abs_error: number;
	    best_epoch: number;
	    weights_path: string;
	    model_name: string;
	    accepted: boolean;
	    validation_games: number;
	    validation_win_rate: number;
	    league_size: number;
	    champion_model: string;
	    champion_score: number;
	
	    static createFrom(source: any = {}) {
	        return new TrainSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.examples = source["examples"];
	        this.train_examples = source["train_examples"];
	        this.validation_examples = source["validation_examples"];
	        this.epochs = source["epochs"];
	        this.learning_rate = source["learning_rate"];
	        this.l2 = source["l2"];
	        this.avg_abs_error = source["avg_abs_error"];
	        this.validation_abs_error = source["validation_abs_error"];
	        this.best_epoch = source["best_epoch"];
	        this.weights_path = source["weights_path"];
	        this.model_name = source["model_name"];
	        this.accepted = source["accepted"];
	        this.validation_games = source["validation_games"];
	        this.validation_win_rate = source["validation_win_rate"];
	        this.league_size = source["league_size"];
	        this.champion_model = source["champion_model"];
	        this.champion_score = source["champion_score"];
	    }
	}

}

