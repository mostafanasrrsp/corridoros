from tactile_power import size_contact_array, PRESETS, precharge_resistor

def test_contact_array():
    res = size_contact_array(30.0, PRESETS['pogo'])
    assert res['contacts_parallel'] >= 6

def test_precharge():
    R, tau, t90 = precharge_resistor(48.0, 0.01, 3.0)
    assert R > 0 and tau > 0 and t90 > 0
