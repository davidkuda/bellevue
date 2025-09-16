window.addEventListener("htmx:load", (e) => loginSpinner(e.target));

function loginSpinner(tree = document) {
	const form = tree.querySelector("#login-form");
	const button = tree.querySelector("#login-button");

	if (!form || !button) {
		return;
	}

	form.addEventListener("submit", () => {
		button.setAttribute("aria-busy", "true");
		button.disabled = true;
		button.classList.add("loading");
	});

	form.addEventListener("htmx:afterSwap", (event) => {
		// Check if an error span got updated (meaning login failed)
		if (form.querySelector("span.error")?.textContent.trim()) {
			button.removeAttribute("aria-busy");
			button.disabled = false;
			button.classList.remove("loading");
		}
	});
}
