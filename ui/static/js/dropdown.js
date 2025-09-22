window.addEventListener("htmx:load", (e) => overflowMenu(e.target));

function overflowMenu(tree = document) {
	tree.querySelectorAll("[data-overflow-menu]").forEach((menuRoot) => {
		const button = menuRoot.querySelector("[aria-haspopup]"),
			menu = menuRoot.querySelector("[role=menu]"),
			items = [...menu.querySelectorAll("[role=menuitem]")];
		const isOpen = () => !menu.hidden;
		items.forEach((item) => item.setAttribute("tabindex", "-1"));
		toggleMenu(isOpen());
		button.addEventListener("click", () => toggleMenu());
		menuRoot.addEventListener("blur", (e) => toggleMenu(false));

		function toggleMenu(open = !isOpen()) {
			if (open) {
				menu.hidden = false;
				button.setAttribute("aria-expanded", "true");
				items[0].focus();
			} else {
				menu.hidden = true;
				button.setAttribute("aria-expanded", "false");
			}
		}

		window.addEventListener("click", function clickAway(event) {
			if (!menuRoot.isConnected) window.removeEventListener("click", clickAway);
			if (!menuRoot.contains(event.target)) toggleMenu(false);
		});

		const currentIndex = () => {
			const idx = items.indexOf(document.activeElement);
			if (idx === -1) return 0;
			return idx;
		};

		menu.addEventListener("keydown", (e) => {
			if (e.key === "ArrowUp") {
				items[currentIndex() - 1]?.focus();
			} else if (e.key === "ArrowDown") {
				items[currentIndex() + 1]?.focus();
			} else if (e.key === "Space") {
				items[currentIndex()].click();
			} else if (e.key === "Home") {
				items[0].focus();
			} else if (e.key === "End") {
				items[items.length - 1].focus();
			} else if (e.key === "Escape") {
				toggleMenu(false);
				button.focus();
			}
		});
	});
}

window.addEventListener("htmx:load", (e) => toggleInvoiceState(e.target));

function toggleInvoiceState(tree = document) {
	const table = tree.querySelector("[data-toggle-invoice-state]");
	table.addEventListener("htmx:afterRequest", (e) => {
		if (e.detail.successful) {
			const tag = e.detail.target;
			const a = e.target;

			const requestPath = e.detail.pathInfo.requestPath;
			const url = new URL(requestPath, location.origin);
			const state = url.searchParams.get("set-state");
			// TODO: maybe control flow on state instead of classList?

			if (tag.classList.contains("state--open")) {
				tag.classList.replace("state--open", "state--paid");
				tag.innerText = "paid";
				a.innerText = "Set Open";
				url.searchParams.set("set-state", "open");
				a.setAttribute("hx-patch", url.pathname + url.search);
			} else {
				tag.classList.replace("state--paid", "state--open");
				tag.innerText = "open";
				a.innerText = "Set Paid";
				url.searchParams.set("set-state", "paid");
				a.setAttribute("hx-patch", url.pathname + url.search);
			}
		}
	});
}
