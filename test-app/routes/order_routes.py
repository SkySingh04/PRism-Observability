
# routes/order_routes.py
from flask import Blueprint, jsonify, request
from services.order_service import create_order

bp = Blueprint('order', __name__, url_prefix='/orders')

@bp.route('/create', methods=['POST'])
def new_order():
    data = request.json
    return jsonify(create_order(data))

    @bp.route('/<id>', methods=['GET'])
    def get_order(id):
        # Logic to fetch order by id
        return jsonify({"id": id, "status": "retrieved"})

    @bp.route('/<id>', methods=['PUT'])
    def update_order(id):
        # Logic to update order
        return jsonify({"id": id, "status": "updated"})

    @bp.route('/<id>', methods=['DELETE'])
    def delete_order(id):
        # Logic to delete order
        return jsonify({"id": id, "status": "deleted"})