import { Link } from "@tanstack/react-router";
import "./Homepage.css";

export function HomePage() {
    return (

        <main className="min-h-screen flex flex-col items-center justify-center bg-slate-50 relative overflow-hidden">
            {/* Animated wave */}
            <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <svg className="w-full h-32 animate-wave-move" viewBox="0 0 1200 100" preserveAspectRatio="none">
                    <path d="M0,50 Q150,0 300,50 T600,50 T900,50 T1200,50" fill="none" stroke="#e2e8f0" strokeWidth="2" />
                    <path
                        className="animate-wave-glow"
                        d="M0,50 Q150,0 300,50 T600,50 T900,50 T1200,50"
                        fill="none"
                        stroke="#1e293b"
                        strokeWidth="3"
                        strokeDasharray="100 1100"
                        strokeLinecap="round"
                    />
                </svg>
            </div>

            {/* Content */}
            <div className="z-10 flex flex-col items-center gap-8 text-center px-4">
                <h1 className="text-4xl md:text-6xl font-bold text-slate-900 tracking-tight">Welcome</h1>
                <p className="text-slate-500 text-lg max-w-md">Get started by signing in to your account</p>
                <Link
                    to="/signin"
                    className="px-8 py-3 bg-slate-900 text-white font-medium rounded-lg hover:bg-slate-800 transition-colors focus:outline-none focus:ring-2 focus:ring-slate-900 focus:ring-offset-2"
                >
                    Sign In
                </Link>

            </div>
        </main>
    )
}

export default HomePage