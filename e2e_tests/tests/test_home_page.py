import re
from playwright.sync_api import Page, expect

def test_has_title(page: Page):
    page.goto("http://localhost:8875/")

    # Expect a title "to contain" a substring.
    expect(page).to_have_title(re.compile("Bellevue Team Activities"))