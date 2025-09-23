from dataclasses import dataclass
from typing import Tuple
import math

@dataclass
class IPTInputs:
    power_W: float
    freq_Hz: float
    primary_L_uH: float
    secondary_L_uH: float
    coupling_k: float
    dc_link_V: float

@dataclass
class LCCCompensation:
    Cs_primary_uF: float
    Cp_secondary_uF: float
    Z_in_ohm: float
    notes: str

def lcc_series_parallel_design(inp: IPTInputs) -> LCCCompensation:
    w = 2 * math.pi * inp.freq_Hz
    Lp = inp.primary_L_uH * 1e-6
    Ls = inp.secondary_L_uH * 1e-6
    L_leak_p = Lp * (1 - inp.coupling_k**2)
    L_leak_s = Ls * (1 - inp.coupling_k**2)
    Cs = 1.0 / (w**2 * L_leak_p) if L_leak_p > 0 else 0.0
    Cp = 1.0 / (w**2 * L_leak_s) if L_leak_s > 0 else 0.0
    V_rms = inp.dc_link_V / math.sqrt(2.0)
    Z_in = (V_rms**2) / max(1e-6, inp.power_W)
    return LCCCompensation(Cs_primary_uF=Cs*1e6, Cp_secondary_uF=Cp*1e6, Z_in_ohm=Z_in,
                           notes="Heuristic LCC sizing — validate with full magnetic model.")
