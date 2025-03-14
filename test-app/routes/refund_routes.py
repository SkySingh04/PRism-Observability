from flask import Blueprint, request, jsonify

refund_bp = Blueprint('refund', __name__, url_prefix='/refund')

@refund_bp.route('/', methods=['POST'])
def create_refund():
    data = request.get_json()
    # Process the refund creation logic here
    # Example:
    order_id = data.get('order_id')
    amount = data.get('amount')

    if not order_id or not amount:
        return jsonify({'error': 'Order ID and amount are required'}), 400

    # In a real application, you would interact with a database or payment gateway here
    # For this example, we'll just return a success message
    return jsonify({'message': f'Refund of {amount} initiated for order {order_id}'}), 201

@refund_bp.route('/<refund_id>', methods=['GET'])
def get_refund(refund_id):
    # Retrieve refund details by ID
    # In a real application, you would query a database
    return jsonify({'refund_id': refund_id, 'status': 'pending', 'amount': 10.00}), 200

@refund_bp.route('/<refund_id>', methods=['PUT'])
def update_refund(refund_id):
    data = request.get_json()
    # Update refund details
    # In a real application, you would update a database record
    status = data.get('status')
    return jsonify({'refund_id': refund_id, 'new_status': status}), 200

@refund_bp.route('/<refund_id>', methods=['DELETE'])
def delete_refund(refund_id):
    # Delete a refund
    # In a real application, you would delete a database record
    return jsonify({'message': f'Refund {refund_id} deleted'}), 200