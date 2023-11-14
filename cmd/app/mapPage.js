var map = L.map("map").setView([34.4208, -119.6982], 13);

L.tileLayer(
    "https://tile.jawg.io/jawg-terrain/{z}/{x}/{y}{r}.png?access-token=LxxdRh2WgA3b9Cs6Zo2HvwYIOxmi9mso32RRvHI1ag8OJHfvICCekhFPKtDxrOC1",
    {}
  ).addTo(map);
