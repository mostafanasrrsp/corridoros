from dataclasses import dataclass
from typing import Tuple, Optional
import math

@dataclass
class ContactType:
    name: str
    max_continuous_current: float
    resistance_milliohm: float
    thermal_rise_C_per_W: float = 10.0

PRESETS = {
    'pogo': ContactType('pogo', max_continuous_current=7.5, resistance_milliohm=20.0, thermal_rise_C_per_W=18.0),
    'multilam': ContactType('multilam', max_continuous_current=60.0, resistance_milliohm=0.4, thermal_rise_C_per_W=6.0),
    'sb120': ContactType('sb120', max_continuous_current=120.0, resistance_milliohm=0.25, thermal_rise_C_per_W=5.0),
}

def contacts_needed(total_current: float, contact: ContactType, utilization: float = 0.7) -> int:
    per = contact.max_continuous_current * utilization
    return max(1, math.ceil(total_current / per))

def array_resistance_milliohm(contact: ContactType, n_parallel: int) -> float:
    return contact.resistance_milliohm / n_parallel

def voltage_drop(total_current: float, r_milliohm: float) -> float:
    return total_current * (r_milliohm / 1000.0)

def power_dissipation(total_current: float, r_milliohm: float) -> float:
    r = r_milliohm / 1000.0
    return total_current * total_current * r

def thermal_rise_C(power_W: float, contact: ContactType) -> float:
    return power_W * contact.thermal_rise_C_per_W

def size_contact_array(total_current: float, contact: ContactType, utilization: float = 0.7):
    n = contacts_needed(total_current, contact, utilization)
    r_milliohm = array_resistance_milliohm(contact, n)
    vdrop = voltage_drop(total_current, r_milliohm)
    p = power_dissipation(total_current, r_milliohm)
    dt = thermal_rise_C(p, contact)
    return {
        'contacts_parallel': n,
        'effective_resistance_milliohm': r_milliohm,
        'voltage_drop_V': vdrop,
        'power_W': p,
        'thermal_rise_C': dt,
    }

def precharge_resistor(bus_voltage: float, bus_capacitance_F: float, inrush_limit_A: float) -> Tuple[float, float, float]:
    R = bus_voltage / max(1e-12, inrush_limit_A)
    tau = R * bus_capacitance_F
    t90 = 2.302585093 * tau
    return R, tau, t90

def precharge_time_to_fraction(R_ohm: float, C_F: float, fraction: float = 0.95) -> float:
    fraction = min(max(fraction, 0.0), 0.999999)
    tau = R_ohm * C_F
    return -math.log(1.0 - fraction) * tau

def inrush_at_time(bus_voltage: float, R_ohm: float, C_F: float, t_s: float) -> float:
    return (bus_voltage / R_ohm) * math.exp(-t_s / (R_ohm * C_F))
