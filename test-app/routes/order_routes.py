
# routes/order_routes.py
from flask import Blueprint, jsonify, request
from services.order_service import create_order

bp = Blueprint('order', __name__, url_prefix='/orders')

@bp.route('/create', methods=['POST'])
def new_order():
    data = request.json
    return jsonify(create_order(data))
