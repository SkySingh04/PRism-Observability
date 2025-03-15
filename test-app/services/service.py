

# services/user_service.py
def get_users():
    return [{"id": 1, "name": "John Doe"}, {"id": 2, "name": "Jane Doe"}]

# services/product_service.py
def get_products():
    return [{"id": 1, "name": "Laptop"}, {"id": 2, "name": "Phone"}]

# services/order_service.py
def create_order(data):
    return {"message": "Order created", "order": data}
