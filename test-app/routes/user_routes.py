# routes/user_routes.py
from flask import Blueprint, jsonify, request
from services.user_service import get_users, create_user, get_user, update_user, delete_user

bp = Blueprint('user', __name__, url_prefix='/users')

@bp.route('/', methods=['GET'])
def list_users():
    return jsonify(get_users())

@bp.route('/<int:id>', methods=['GET'])
def get_user_by_id(id):
    user = get_user(id)
    if user:
        return jsonify(user)
    return jsonify({'message': 'User not found'}), 404

@bp.route('/', methods=['POST'])
def create_new_user():
    data = request.get_json()
    user = create_user(data)
    return jsonify(user), 201

@bp.route('/<int:id>', methods=['PUT'])
def update_existing_user(id):
    data = request.get_json()
    user = update_user(id, data)
    if user:
        return jsonify(user)
    return jsonify({'message': 'User not found'}), 404

@bp.route('/<int:id>', methods=['DELETE'])
def delete_existing_user(id):
    if delete_user(id):
        return jsonify({'message': 'User deleted'})
    return jsonify({'message': 'User not found'}), 404