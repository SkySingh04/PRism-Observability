from flask import Blueprint, jsonify

return_routes = Blueprint('return_routes', __name__)

@return_routes.route('/return/success', methods=['GET'])
def return_success():
    return jsonify({
        'status': 'success',
        'message': 'Operation completed successfully'
    }), 200

@return_routes.route('/return/error', methods=['GET'])
def return_error():
    return jsonify({
        'status': 'error',
        'message': 'An error occurred'
    }), 500

@return_routes.route('/return/not-found', methods=['GET'])
def return_not_found():
    return jsonify({
        'status': 'error',
        'message': 'Resource not found'
    }), 404

@return_routes.route('/return/bad-request', methods=['GET'])
def return_bad_request():
    return jsonify({
        'status': 'error',
        'message': 'Bad request'
    }), 400