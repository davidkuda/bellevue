# conftest.py

import pytest

@pytest.fixture(scope="class")
def context(browser):
    """ Overwrites the default context scope to be 'class'. 
    This ensures all tests in a class use the same browser context/session.
    I'm not sure this is the correct way to do this in playwright as I'm new to this library.
    """
    context = browser.new_context()
    yield context
    context.close()

@pytest.fixture(scope="class")
def page(context):
    """ Creates a page object scoped to the class context."""
    page = context.new_page()
    return page