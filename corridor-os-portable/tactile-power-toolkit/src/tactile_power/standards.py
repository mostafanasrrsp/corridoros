from dataclasses import dataclass
from typing import Tuple

@dataclass
class CreepageClearanceInputs:
    rated_voltage: float
    pollution_degree: int
    material_group: str
    overvoltage_category: int = 2
    altitude_m: int = 2000
    safety_factor: float = 1.25

def recommended_clearance_mm(v: float, ov_cat: int = 2) -> float:
    if v <= 50: base = 1.0
    elif v <= 150: base = 1.5
    elif v <= 300: base = 2.5
    elif v <= 600: base = 5.0
    else: base = 8.0
    bump = 1.0 + max(0, ov_cat - 2) * 0.2
    return base * bump

def creepage_per_voltage(v: float, pollution_degree: int, material_group: str) -> float:
    mg = material_group.upper()
    if pollution_degree <= 1:
        mm_per_100v = 0.8 if mg == 'I' else 1.0
    elif pollution_degree == 2:
        mm_per_100v = 2.5 if mg in ('I', 'II') else 3.2
    elif pollution_degree == 3:
        mm_per_100v = 4.0 if mg in ('I', 'II') else 5.0
    else:
        mm_per_100v = 8.0
    return (v / 100.0) * mm_per_100v

def recommend_creepage_clearance(inputs: CreepageClearanceInputs) -> Tuple[float, float]:
    clear_mm = recommended_clearance_mm(inputs.rated_voltage, inputs.overvoltage_category)
    creep_mm = creepage_per_voltage(inputs.rated_voltage, inputs.pollution_degree, inputs.material_group)
    if inputs.altitude_m > 2000:
        extra = 0.1 * ((inputs.altitude_m - 2000) / 1000.0)
        clear_mm *= (1.0 + extra)
    return clear_mm * inputs.safety_factor, creep_mm * inputs.safety_factor
