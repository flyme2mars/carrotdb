import time

from carrotdb import Client


def run_example():
    print("🥕 CarrotDB Python SDK Example")

    try:
        # Connect to 'billing' database
        billing_db = Client(host="localhost", port=8000, database="billing")
        billing_db.set("invoice_1", "paid")
        print(f"[Billing] invoice_1: {billing_db.get('invoice_1')}")

        # Connect to 'users' database
        users_db = Client(host="localhost", port=8000, database="users")
        users_db.set("user_1", "Joe")
        print(f"[Users] user_1: {users_db.get('user_1')}")

        # Verify Isolation
        print(
            f"[Billing] Checking user_1 in billing: {billing_db.get('user_1')} (Expected: None)"
        )

        # List keys
        print(f"[Users] Keys in users db: {users_db.list_keys()}")
        print(f"[Billing] Keys in billing db: {billing_db.list_keys()}")

    except Exception as e:
        print(f"Error: {e}")
        print("Tip: Make sure CarrotDB Router is running on port 8000!")


if __name__ == "__main__":
    run_example()
