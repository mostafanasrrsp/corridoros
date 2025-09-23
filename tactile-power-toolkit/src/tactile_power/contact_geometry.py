from dataclasses import dataclass
from typing import Tuple, List
import math

@dataclass
class TileSpec:
    pad_size_mm: float
    pad_gap_mm: float
    skirt_height_mm: float
    magnet_diameter_mm: float
    magnet_offset_mm: float
    tile_size_mm: float
    creepage_lane_mm: float

def finger_safe_ip2x_ok(opening_diameter_mm: float) -> bool:
    return opening_diameter_mm <= 12.3

def magnet_positions(spec: TileSpec):
    d = spec.magnet_offset_mm
    return [(-d, -d), (d, -d), (d, d), (-d, d)]

def pad_positions(spec: TileSpec):
    s = spec.pad_size_mm/2.0 + spec.creepage_lane_mm/2.0
    return [(-s, 0.0, '+'), (s, 0.0, '-')]

def required_creepage_lane_for_voltage(v: float, mm_per_100V: float = 3.2, safety_factor: float = 1.25) -> float:
    return (v/100.0) * mm_per_100V * safety_factor

def symmetric_tile_for_voltage(v: float, current_A: float) -> TileSpec:
    pad = 18.0 + 0.05 * max(0.0, current_A - 20.0)
    lane = required_creepage_lane_for_voltage(v)
    gap = max(4.0, lane)
    magnet_d = 8.0 if current_A <= 60 else 10.0
    tile_size = max(2*(pad + gap) + 12.0, 60.0)
    return TileSpec(
        pad_size_mm=pad,
        pad_gap_mm=gap,
        skirt_height_mm=3.0,
        magnet_diameter_mm=magnet_d,
        magnet_offset_mm=0.35 * tile_size/2.0,
        tile_size_mm=tile_size,
        creepage_lane_mm=lane,
    )
