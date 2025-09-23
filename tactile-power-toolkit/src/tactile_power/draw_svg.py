from .contact_geometry import TileSpec, magnet_positions, pad_positions

def _rect(xc, yc, w, h, rx=1.5, fill="#ddd", stroke="#000"):
    x = xc - w/2.0
    y = yc - h/2.0
    return f'<rect x="{x:.2f}" y="{y:.2f}" width="{w:.2f}" height="{h:.2f}" rx="{rx}" fill="{fill}" stroke="{stroke}" stroke-width="0.8"/>'

def _circle(xc, yc, r, fill="#bbb", stroke="#000"):
    return f'<circle cx="{xc:.2f}" cy="{yc:.2f}" r="{r:.2f}" fill="{fill}" stroke="{stroke}" stroke-width="0.8"/>'

def tile_svg(spec: TileSpec, unit_mm: float = 1.0) -> str:
    s = spec.tile_size_mm
    w = s; h = s
    svg = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{w}mm" height="{h}mm" viewBox="{-s/2} {-s/2} {s} {s}">']
    svg.append(_rect(0, 0, s, s, rx=4, fill="#f7f7f7"))
    pad_w = spec.pad_size_mm; pad_h = spec.pad_size_mm
    for x, y, pol in pad_positions(spec):
        color = "#e0e0ff" if pol == "+" else "#ffe0e0"
        svg.append(_rect(x, y, pad_w, pad_h, rx=2, fill=color))
        svg.append(f'<text x="{x:.2f}" y="{y + pad_h/2 + 4:.2f}" font-size="4" text-anchor="middle">{pol}</text>')
    svg.append(_rect(0, 0, spec.pad_gap_mm + 0.1, s*0.8, rx=0, fill="#fafafa", stroke="#aaa"))
    for (mx, my) in magnet_positions(spec):
        svg.append(_circle(mx, my, spec.magnet_diameter_mm/2.0, fill="#c0c0c0"))
    svg.append(f'<text x="0" y="{-s/2 + 6:.2f}" font-size="4" text-anchor="middle">skirt {spec.skirt_height_mm} mm</text>')
    svg.append('</svg>')
    return "\n".join(svg)
