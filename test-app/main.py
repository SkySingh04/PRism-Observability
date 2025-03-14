# app.py
from flask import Flask
from routes import user_routes, product_routes, order_routes

def create_app():
    app = Flask(__name__)
    
    app.register_blueprint(user_routes.bp)
    app.register_blueprint(product_routes.bp)
    app.register_blueprint(order_routes.bp)
    
    return app

if __name__ == "__main__":
    app = create_app()
    app.run(debug=True)
