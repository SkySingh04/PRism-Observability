from flask import Blueprint, request, jsonify

remittance_bp = Blueprint('remittance', __name__, url_prefix='/remittance')

@remittance_bp.route('/send', methods=['POST'])
def send_remittance():
    data = request.get_json()
    # Process remittance sending logic here
    return jsonify({'message': 'Remittance sent successfully'}), 200

@remittance_bp.route('/status/<transaction_id>', methods=['GET'])
def get_remittance_status(transaction_id):
    # Retrieve and return remittance status based on transaction_id
    return jsonify({'status': 'pending'}), 200

@remittance_bp.route('/cancel/<transaction_id>', methods=['POST'])
def cancel_remittance(transaction_id):
    # Logic to cancel a remittance
    return jsonify({'message': f'Remittance {transaction_id} cancelled'}), 200