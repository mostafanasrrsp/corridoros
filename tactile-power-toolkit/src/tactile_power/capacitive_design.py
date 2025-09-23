from dataclasses import dataclass
import math

@dataclass
class CPTInputs:
    power_W: float
    freq_Hz: float
    gap_mm: float
    epsilon_r: float
    v_rms: float

@dataclass
class CPTDesign:
    cap_uF: float
    plate_area_cm2: float
    reactive_VA: float
    notes: str

def design_cpt(inp: CPTInputs) -> CPTDesign:
    w = 2 * math.pi * inp.freq_Hz
    Z_target = (inp.v_rms**2) / max(1e-6, inp.power_W)
    C = 1.0 / (w * Z_target)
    eps0 = 8.854e-12
    d = inp.gap_mm * 1e-3
    A = C * d / (eps0 * inp.epsilon_r)
    Xc = 1.0 / (w * C)
    Q = (inp.v_rms**2) / Xc
    return CPTDesign(cap_uF=C*1e6, plate_area_cm2=A*1e4, reactive_VA=Q,
                     notes="Parallel-plate estimate; real designs need fringing and guard ring analysis.")
