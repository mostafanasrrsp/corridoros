from tactile_power import size_contact_array, PRESETS, symmetric_tile_for_voltage
from tactile_power import tile_svg, precharge_resistor, recommend_creepage_clearance, CreepageClearanceInputs
from pathlib import Path

def main():
    s = size_contact_array(60.0, PRESETS['pogo'])
    print('Contact array sizing:', s)
    spec = symmetric_tile_for_voltage(48.0, 60.0)
    svg = tile_svg(spec)
    Path('tile_48V_60A.svg').write_text(svg)
    print('Wrote tile_48V_60A.svg')
    R, tau, t90 = precharge_resistor(48.0, 0.01, 5.0)
    print('Precharge:', {'R_ohm': R, 'tau_s': tau, 't90_s': t90})
    cc = recommend_creepage_clearance(CreepageClearanceInputs(48.0, 2, 'II'))
    print('Clearance/Creepage (mm):', cc)

if __name__ == '__main__':
    main()
