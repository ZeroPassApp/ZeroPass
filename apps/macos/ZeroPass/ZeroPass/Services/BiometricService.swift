import Foundation
import LocalAuthentication

enum BiometricError: LocalizedError {
    case unavailable
    case failed(message: String)

    var errorDescription: String? {
        switch self {
        case .unavailable:
            return "Biometric authentication is unavailable"
        case .failed(let message):
            return message
        }
    }
}

nonisolated final class BiometricService {
    func isAvailable() -> Bool {
        var err: NSError?
        return LAContext().canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &err)
    }

    func authenticateContext(reason: String) async throws -> LAContext {
        let ctx = LAContext()
        ctx.localizedCancelTitle = "Cancel"

        var err: NSError?
        guard ctx.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &err) else {
            throw BiometricError.unavailable
        }

        try await withCheckedThrowingContinuation { (cont: CheckedContinuation<Void, Error>) in
            ctx.evaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, localizedReason: reason) { ok, error in
                if ok {
                    cont.resume()
                } else {
                    let msg = (error as NSError?)?.localizedDescription ?? "Authentication failed"
                    cont.resume(throwing: BiometricError.failed(message: msg))
                }
            }
        }

        return ctx
    }

    func authenticate(reason: String) async throws {
        _ = try await authenticateContext(reason: reason)
    }
}
