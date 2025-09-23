import argparse, json, sys
from .electrical_sizing import PRESETS, ContactType, size_contact_array, precharge_resistor
from .standards import CreepageClearanceInputs, recommend_creepage_clearance
from .contact_geometry import symmetric_tile_for_voltage
from .draw_svg import tile_svg
from .inductive_design import IPTInputs, lcc_series_parallel_design
from .capacitive_design import CPTInputs, design_cpt

def cmd_size_contacts(args):
    ct = PRESETS.get(args.contact_type, ContactType(args.contact_type, args.max_per_contact, args.contact_resistance_milliohm))
    if args.max_per_contact:
        ct.max_continuous_current = args.max_per_contact
    if args.contact_resistance_milliohm:
        ct.resistance_milliohm = args.contact_resistance_milliohm
    data = size_contact_array(args.current, ct, args.utilization)
    print(json.dumps(data, indent=2))

def cmd_precharge(args):
    R, tau, t90 = precharge_resistor(args.voltage, args.capacitance, args.inrush_limit)
    out = {"R_ohm": R, "tau_s": tau, "t90_s": t90}
    print(json.dumps(out, indent=2))

def cmd_clearances(args):
    cc = recommend_creepage_clearance(CreepageClearanceInputs(rated_voltage=args.voltage,
                                                              pollution_degree=args.pollution_degree,
                                                              material_group=args.material_group,
                                                              overvoltage_category=args.ov_cat,
                                                              altitude_m=args.altitude_m,
                                                              safety_factor=args.safety_factor))
    print(json.dumps({"clearance_mm": cc[0], "creepage_mm": cc[1]}, indent=2))

def cmd_tile_svg(args):
    spec = symmetric_tile_for_voltage(args.voltage, args.current)
    svg = tile_svg(spec)
    with open(args.outfile, 'w') as f:
        f.write(svg)
    print(json.dumps({"outfile": args.outfile,
                      "tile_size_mm": spec.tile_size_mm,
                      "pad_size_mm": spec.pad_size_mm,
                      "creepage_lane_mm": spec.creepage_lane_mm}, indent=2))

def cmd_inductive(args):
    inp = IPTInputs(power_W=args.power, freq_Hz=args.freq, primary_L_uH=args.Lp_uH,
                    secondary_L_uH=args.Ls_uH, coupling_k=args.k, dc_link_V=args.dc)
    out = lcc_series_parallel_design(inp)
    print(json.dumps({"Cs_primary_uF": out.Cs_primary_uF,
                      "Cp_secondary_uF": out.Cp_secondary_uF,
                      "Z_in_ohm": out.Z_in_ohm,
                      "notes": out.notes}, indent=2))

def cmd_capacitive(args):
    inp = CPTInputs(power_W=args.power, freq_Hz=args.freq, gap_mm=args.gap, epsilon_r=args.er, v_rms=args.vrms)
    out = design_cpt(inp)
    print(json.dumps({"cap_uF": out.cap_uF,
                      "plate_area_cm2": out.plate_area_cm2,
                      "reactive_VA": out.reactive_VA,
                      "notes": out.notes}, indent=2))

def main():
    p = argparse.ArgumentParser(prog='tactile-power', description='Tactile power design CLI')
    sub = p.add_subparsers(dest='cmd')

    s1 = sub.add_parser('size-contacts', help='Size parallel contacts for current')
    s1.add_argument('--current', type=float, required=True, help='Total current (A)')
    s1.add_argument('--contact-type', type=str, default='pogo', help='pogo|multilam|sb120 or custom')
    s1.add_argument('--max-per-contact', type=float, default=0.0, help='Override preset (A/contact)')
    s1.add_argument('--contact-resistance-milliohm', type=float, default=0.0, help='Override preset (mΩ/contact)')
    s1.add_argument('--utilization', type=float, default=0.7, help='Fraction of rated current per contact')
    s1.set_defaults(func=cmd_size_contacts)

    s2 = sub.add_parser('precharge', help='Compute pre-charge resistor and timing')
    s2.add_argument('--voltage', type=float, required=True)
    s2.add_argument('--capacitance', type=float, required=True, help='Bus capacitance (F)')
    s2.add_argument('--inrush-limit', type=float, required=True, help='Limit current (A)')
    s2.set_defaults(func=cmd_precharge)

    s3 = sub.add_parser('clearances', help='Recommend creepage/clearance (heuristic)')
    s3.add_argument('--voltage', type=float, required=True)
    s3.add_argument('--pollution-degree', type=int, default=2)
    s3.add_argument('--material-group', type=str, default='II')
    s3.add_argument('--ov-cat', type=int, default=2)
    s3.add_argument('--altitude-m', type=int, default=2000)
    s3.add_argument('--safety-factor', type=float, default=1.25)
    s3.set_defaults(func=cmd_clearances)

    s4 = sub.add_parser('tile-svg', help='Generate a tile SVG')
    s4.add_argument('--outfile', type=str, required=True)
    s4.add_argument('--voltage', type=float, required=True)
    s4.add_argument('--current', type=float, required=True)
    s4.add_argument('--pad-size', type=float, default=0.0)
    s4.add_argument('--pad-gap', type=float, default=0.0)
    s4.add_argument('--magnet-d', type=float, default=0.0)
    s4.set_defaults(func=cmd_tile_svg)

    s5 = sub.add_parser('inductive', help='Rough LCC compensation for IPT')
    s5.add_argument('--power', type=float, required=True)
    s5.add_argument('--freq', type=float, required=True)
    s5.add_argument('--Lp-uH', dest='Lp_uH', type=float, required=True)
    s5.add_argument('--Ls-uH', dest='Ls_uH', type=float, required=True)
    s5.add_argument('--k', type=float, required=True, help='Coupling coefficient (0..1)')
    s5.add_argument('--dc', type=float, required=True, help='DC link voltage (V)')
    s5.set_defaults(func=cmd_inductive)

    s6 = sub.add_parser('capacitive', help='Capacitive plate sizing')
    s6.add_argument('--power', type=float, required=True)
    s6.add_argument('--freq', type=float, required=True)
    s6.add_argument('--gap', type=float, required=True, help='Gap (mm)')
    s6.add_argument('--er', type=float, required=True, help='Relative permittivity')
    s6.add_argument('--vrms', type=float, required=True, help='Drive voltage RMS')
    s6.set_defaults(func=cmd_capacitive)

    args = p.parse_args()
    if not hasattr(args, 'func'):
        p.print_help(); import sys; sys.exit(1)
    args.func(args)

if __name__ == '__main__':
    main()
