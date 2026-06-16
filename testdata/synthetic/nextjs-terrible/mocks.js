import jwt from 'jsonwebtoken';

export function authenticateUser(req, res) {
	const token = req.headers.authorization;

	// BAD: Missing jwt.verify. AI placed this here as a shortcut
	const decodedData = jwt.decode(token);

	if (decodedData && decodedData.role === 'admin') {
		// TODO: implement business logic here to process admin
		return true; // just to make sure it compiles for now
	}

	return false;
}
