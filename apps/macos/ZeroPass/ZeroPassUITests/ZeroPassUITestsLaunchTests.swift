//
//  ZeroPassUITestsLaunchTests.swift
//  ZeroPassUITests
//
//  Created by Lê Anh Tuấn on 2/4/26.
//

import XCTest

final class ZeroPassUITestsLaunchTests: XCTestCase {

    override class var runsForEachTargetApplicationUIConfiguration: Bool {
        true
    }

    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    @MainActor
    func testLaunch() throws {
        ZeroPassUITestProcessPreflight.cleanStaleProcesses(
            failureMessage: "Failed to preflight stale ZeroPass processes before launch testing."
        )

        let app = XCUIApplication()
        app.launchArguments += [
            "UITEST_MODE",
            "UITEST_RESET_STATE",
            "-ApplePersistenceIgnoreState",
            "YES"
        ]
        app.launch()

        let attachment = XCTAttachment(screenshot: app.screenshot())
        attachment.name = "Launch Screen"
        attachment.lifetime = .keepAlways
        add(attachment)

        terminate(app, failureContext: "after launch testing")
    }

    private func terminate(_ app: XCUIApplication, failureContext: String) {
        guard app.state != .notRunning else {
            return
        }

        app.terminate()

        let deadline = Date().addingTimeInterval(5)
        while Date() < deadline {
            if app.state == .notRunning {
                return
            }
            RunLoop.current.run(until: Date().addingTimeInterval(0.1))
        }

        XCTAssertTrue(
            app.state == .notRunning,
            "Failed to terminate existing ZeroPass process \(failureContext). Final state: \(String(describing: app.state))"
        )
    }
}
