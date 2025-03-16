# routes/order_routes.py
from flask import Blueprint, jsonify, request
from services.order_service import create_order, get_order_by_id, update_order_by_id, delete_order_by_id

bp = Blueprint('order', __name__, url_prefix='/orders')

@bp.route('/create', methods=['POST'])
def new_order():
    data = request.json
    return jsonify(create_order(data))

@bp.route('/<id>', methods=['GET'])
def get_order(id):
    order = get_order_by_id(id)
    if order:
        return jsonify(order)
    return jsonify({"message": "Order not found"}), 404

@bp.route('/<id>', methods=['PUT'])
def update_order(id):
    data = request.json
    updated_order = update_order_by_id(id, data)
    if updated_order:
        return jsonify(updated_order)
    return jsonify({"message": "Order not found"}), 404

@bp.route('/<id>', methods=['DELETE'])
def delete_order(id):
    if delete_order_by_id(id):
        return jsonify({"message": "Order deleted"})
    return jsonify({"message": "Order not found"}), 404