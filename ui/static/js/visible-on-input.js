window.addEventListener("htmx:load", (e) =>
	togglePriceCategoryVisibility(e.target)
);

function togglePriceCategoryVisibility(tree = document) {
	tree.querySelectorAll("fieldset.fixed-price-activity").forEach((fieldset) => {
		const counter = fieldset.querySelector("input[type='number']");
		counter.addEventListener("input", (e) => {
			const count = e.target.value;
			const priceCategory = fieldset.querySelector("fieldset.price-categories");
			if (count > 0) {
				priceCategory.style.display = "inline-block";
			} else {
				priceCategory.style.display = "none";
			}
		});
	});
}
