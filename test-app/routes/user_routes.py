# routes/user_routes.py
from flask import Blueprint, jsonify
from services.user_service import get_users

bp = Blueprint('user', __name__, url_prefix='/users')

@bp.route('/', methods=['GET'])
def list_users():
    return jsonify(get_users())