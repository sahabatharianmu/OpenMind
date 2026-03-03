import { createRoot } from "react-dom/client";
import App from "./App.tsx";
import "./lib/i18n"; // Initialize i18n before React
import "./index.css";

createRoot(document.getElementById("root")!).render(<App />);
