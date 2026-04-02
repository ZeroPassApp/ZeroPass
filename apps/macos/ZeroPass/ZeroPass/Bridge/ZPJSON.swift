import Foundation

enum ZPJSON {
    static let decoder: JSONDecoder = {
        let d = JSONDecoder()

        d.dateDecodingStrategy = .custom { decoder in
            let c = try decoder.singleValueContainer()
            let s = try c.decode(String.self)

            let iso = ISO8601DateFormatter()
            iso.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
            if let dt = iso.date(from: s) { return dt }

            let isoNoFrac = ISO8601DateFormatter()
            isoNoFrac.formatOptions = [.withInternetDateTime]
            if let dt = isoNoFrac.date(from: s) { return dt }

            throw DecodingError.dataCorruptedError(in: c, debugDescription: "Invalid RFC3339 date: \(s)")
        }

        return d
    }()

    static let encoder: JSONEncoder = {
        let e = JSONEncoder()

        let iso = ISO8601DateFormatter()
        iso.formatOptions = [.withInternetDateTime, .withFractionalSeconds]

        e.dateEncodingStrategy = .custom { date, encoder in
            var c = encoder.singleValueContainer()
            try c.encode(iso.string(from: date))
        }

        e.outputFormatting = [.withoutEscapingSlashes]
        return e
    }()
}
