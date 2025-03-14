
# routes/product_routes.py
from flask import Blueprint, jsonify
from services.product_service import get_products

bp = Blueprint('product', __name__, url_prefix='/products')

@bp.route('/', methods=['GET'])
def list_products():
    return jsonify(get_products())
