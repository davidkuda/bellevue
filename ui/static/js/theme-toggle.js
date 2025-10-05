window.addEventListener("htmx:load", (e) => themeToggle(e.target));

function themeToggle(tree = document) {
	const themeBtn = document.getElementById("themeToggle");

	themeBtn.addEventListener("click", async () => {
		const current =
			document.documentElement.getAttribute("data-theme") || "light";

		// toggle:
		if (current === "light") {
			document.documentElement.setAttribute("data-theme", "dark");
			await localStorage.setItem("theme", "dark");
			themeBtn.textContent = "light mode";
		} else {
			document.documentElement.setAttribute("data-theme", "light");
			await localStorage.setItem("theme", "light");
			themeBtn.textContent = "dark mode";
		}
	});
}

(async () => {
	const current = await localStorage.getItem("theme");
	if (current === "dark") {
		document.getElementById("themeToggle").textContent = "light mode";
	}
})();
