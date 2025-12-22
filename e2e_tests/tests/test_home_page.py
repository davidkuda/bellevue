import re
import time
import random

import pytest
from playwright.sync_api import Page, expect


class TestSequencedSession:
    """ One long test for now as the login we create gives us access to the app itself. 
    TODO: add a test user and split out test of the activity from the test of the login page.
    """
    def test_login_page(self, page: Page):
        page.goto("http://localhost:8875/")

        # Expect a title "to contain" a substring.
        expect(page).to_have_title(re.compile("Bellevue Team Activities"))

        # Check DW Connect login button
        expect(page.get_by_role("button", name="Login with DW-Connect")).to_be_visible()
        expect(page.get_by_role("link", name="Create an account")).to_be_visible()
        expect(page.get_by_role("link", name="login with your email and password")).to_be_visible()

        page.get_by_role("link", name="login with your email and password").click()
        login_form_locator = page.locator('#login-form')
        expect(login_form_locator).to_be_visible()
        expect(page.get_by_role("button", name="Login", exact=True)).to_be_enabled()
        
        specific_inputs_locator = page.locator('#login-form input[type="email"], #login-form input[type="password"]')
        expect(specific_inputs_locator).to_have_count(2)


    def test_register(self, page: Page):
        # TODO: check behaviour when not filling out all fields
        page.get_by_role("link", name="Create an account").click()
        page.fill('input[name="first-name"]', 'Test')
        page.fill('input[name="last-name"]', 'Tester')
        page.fill('input[type="email"]', f'test_{random.randint(10_000, 20_000)}@example.com')
        page.fill('input[type="password"]', 'pa55word')
        page.click('#login-button')
        expect(page.get_by_role("button", name="Add your first activity")).to_be_visible()

    def test_add_first_activity(self, page: Page):
        """ Check we can add an activity """
        page.get_by_role("button", name="Add your first activity").click()

        # Find the container for breakfast and then the plus button within it
        b_plus = page.locator(".fixed-price-activity", has=page.locator("#breakfast-quantity")).locator("button.plus")
        b_plus.click()
        b_plus.click()

        # Submit activity
        page.get_by_role("button", name="Add").click()

        # Now should have table with three rows: header, breakfast we added and total
        expect(page.locator("table tbody tr")).to_have_count(3)


        breakfast_info = page.get_by_text("2 Breakfast for 9.00 CHF (regular)")
        expect(breakfast_info).to_be_visible()
