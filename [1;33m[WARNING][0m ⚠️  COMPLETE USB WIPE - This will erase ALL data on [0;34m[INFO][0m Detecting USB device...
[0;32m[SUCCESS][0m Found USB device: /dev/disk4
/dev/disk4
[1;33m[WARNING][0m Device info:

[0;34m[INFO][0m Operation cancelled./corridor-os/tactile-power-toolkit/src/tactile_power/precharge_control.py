from dataclasses import dataclass
from enum import Enum, auto
from typing import Optional, Callable

class State(Enum):
    IDLE = auto()
    ALIGNING = auto()
    SENSE_MADE = auto()
    PRECHARGING = auto()
    CLOSED = auto()
    OPENING = auto()
    FAULT = auto()

@dataclass
class PrechargeConfig:
    r_pre_ohm: float
    c_bus_F: float
    v_bus: float
    t_min_precharge_s: float
    t_max_precharge_s: float
    i_limit_A: float

@dataclass
class Signals:
    sense: bool = False
    emergency_open: bool = False
    timer_s: float = 0.0
    v_cap: float = 0.0

@dataclass
class Controller:
    cfg: PrechargeConfig
    state: State = State.IDLE
    on_main_close: Optional[Callable[[], None]] = None
    on_main_open: Optional[Callable[[], None]] = None

    def step(self, sig: Signals) -> State:
        if self.state == State.IDLE:
            if sig.sense:
                self.state = State.ALIGNING
        elif self.state == State.ALIGNING:
            if not sig.sense:
                self.state = State.IDLE
            else:
                self.state = State.SENSE_MADE
        elif self.state == State.SENSE_MADE:
            self.state = State.PRECHARGING
        elif self.state == State.PRECHARGING:
            if sig.emergency_open or not sig.sense:
                self.state = State.OPENING
            elif sig.timer_s >= self.cfg.t_min_precharge_s and sig.v_cap >= 0.9 * self.cfg.v_bus:
                if self.on_main_close: self.on_main_close()
                self.state = State.CLOSED
            elif sig.timer_s > self.cfg.t_max_precharge_s:
                self.state = State.FAULT
        elif self.state == State.CLOSED:
            if sig.emergency_open or not sig.sense:
                if self.on_main_open: self.on_main_open()
                self.state = State.OPENING
        elif self.state == State.OPENING:
            self.state = State.IDLE
        elif self.state == State.FAULT:
            if not sig.sense:
                self.state = State.IDLE
        return self.state
