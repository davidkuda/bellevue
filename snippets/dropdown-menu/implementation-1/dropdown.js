document.addEventListener("toggle", e => {
  if (!e.target.matches(".inline-menu")) return;

  const menu = e.target.querySelector("ul");
  if (!menu) return;

  // Reset to default (open to right)
  menu.style.left = "0";
  menu.style.right = "auto";

  if (e.target.open) {
    // Let browser paint, then measure
    requestAnimationFrame(() => {
      const rect = menu.getBoundingClientRect();
      const margin = 8;
      if (rect.right > window.innerWidth - margin) {
        // overflow right → flip to left
        menu.style.left = "auto";
        menu.style.right = "0";
      }
    });
  }
});

