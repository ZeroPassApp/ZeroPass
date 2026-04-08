import React from "react";
import ReactDOM from "react-dom/client";
import "./styles/reset.css";
import "./styles/theme.css";
import "./styles/animations.css";
import App from "./App";

ReactDOM.createRoot(document.getElementById("root") as HTMLElement).render(
    <React.StrictMode>
        <App />
    </React.StrictMode>,
);
