# OpenMind Frontend

The web client for **OpenMind**, an open-source Clinic Management System (CMS) designed for modern healthcare practices. Built with React, TypeScript, and Vite, it delivers a high-performance, responsive experience for clinicians and administrators.

## Features

- **Dashboard**: Real-time overview of active patients, upcoming sessions, note activity, and revenue metrics.
- **Patient Management**: Create, view, and manage patient profiles with rich history tracking.
- **Appointments**: Calendar and list views for scheduling, with support for video and in-person modes.
- **Clinical Notes**: Integrated CKEditor for writing SOAP notes, linked directly to patients and appointments.
- **Billing & Invoicing**: Generate invoices, track payment status, and visualize practice revenue.
- **Authentication**: Secure JWT-based authentication integrated with the OpenMind Go backend.

## Tech Stack

- **Framework**: [React](https://react.dev/) + [Vite](https://vitejs.dev/)
- **Language**: [TypeScript](https://www.typescriptlang.org/)
- **Styling**: [Tailwind CSS](https://tailwindcss.com/) + [Shadcn/UI](https://ui.shadcn.com/)
- **State Management**: [TanStack Query (React Query)](https://tanstack.com/query/latest)
- **Routing**: [React Router](https://reactrouter.com/)
- **Forms**: React Hook Form + Zod validation
- **Icons**: Lucide React
- **Runtime**: [Bun](https://bun.sh/) (recommended) or Node.js

## Prerequisites

- **Bun** (v1.0 or later) or **Node.js** (v18 or later)
- **OpenMind Backend**: The Go backend service must be running locally or accessible via network.

## Installation

1.  **Clone the repository** (if you haven't already):

    ```bash
    git clone https://github.com/sahabatharianmu/OpenMind.git
    cd OpenMind/web
    ```

2.  **Install dependencies**:

    ```bash
    bun install
    # or
    npm install
    ```

3.  **Environment Setup**:
    No strict `.env` setup is required for local dev defaults (proxies to `http://localhost:8080`). To customize the backend URL, update `vite.config.ts`.

## Development

Start the development server with hot-module replacement (HMR):

```bash
bun run dev
# or
npm run dev
```

The application will be available at `http://localhost:5173`.

## Build for Production

To build the application for deployment:

```bash
bun run build
# or
npm run build
```

The output will be in the `dist/` directory, ready to be served by Nginx, Apache, or the Go backend itself.

## Project Structure

```
web/
├── src/
│   ├── api/            # Axios client and interceptors
│   ├── components/     # Reusable UI components (Shadcn, specialized widgets)
│   ├── contexts/       # React Contexts (AuthContext, etc.)
│   ├── hooks/          # Custom Hooks (useDashboardQueries, etc.)
│   ├── pages/          # Application views/routes
│   ├── services/       # API abstraction layer
│   └── types/          # Shared TypeScript interfaces
├── public/
└── vite.config.ts      # Vite configuration & proxy setup
```

## Contributing

We welcome contributions! Please fork the repository and submit a Pull Request.

## License

OpenMind is open-source software.
