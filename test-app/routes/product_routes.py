# routes/product_routes.py
from flask import Blueprint, jsonify, request
from services.product_service import get_products, create_product, get_product, update_product, delete_product

bp = Blueprint('product', __name__, url_prefix='/products')

@bp.route('/', methods=['GET'])
def list_products():
    return jsonify(get_products())

@bp.route('/<id>', methods=['GET'])
def get_product_by_id(id):
    product = get_product(id)
    if product:
        return jsonify(product)
    return jsonify({'message': 'Product not found'}), 404

@bp.route('/', methods=['POST'])
def add_product():
    data = request.get_json()
    product = create_product(data)
    return jsonify(product), 201

@bp.route('/<id>', methods=['PUT'])
def update_product_by_id(id):
    data = request.get_json()
    product = update_product(id, data)
    if product:
        return jsonify(product)
    return jsonify({'message': 'Product not found'}), 404

@bp.route('/<id>', methods=['DELETE'])
def delete_product_by_id(id):
    if delete_product(id):
        return jsonify({'message': 'Product deleted'})
    return jsonify({'message': 'Product not found'}), 404
