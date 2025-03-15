
# models.py
class User:
    def __init__(self, user_id, name):
        self.id = user_id
        self.name = name

class Product:
    def __init__(self, product_id, name):
        self.id = product_id
        self.name = name

class Order:
    def __init__(self, order_id, user_id, product_id):
        self.id = order_id
        self.user_id = user_id
        self.product_id = product_id

