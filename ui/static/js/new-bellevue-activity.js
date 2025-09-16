window.addEventListener("htmx:load", (e) => {
	e.target.querySelectorAll(".number-picker").forEach((picker) => {
		// idempotency guard:
		if (picker.dataset.bound) {
			return;
		}
		picker.dataset.bound = "1";

		const input = picker.querySelector("input");
		picker
			.querySelector(".minus")
			.addEventListener("click", () => input.stepDown());
		picker
			.querySelector(".plus")
			.addEventListener("click", () => input.stepUp());
	});
});
