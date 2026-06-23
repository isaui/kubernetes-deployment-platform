import { ActionFunctionArgs, json, redirect } from "@remix-run/node";
import { Form, Link, useActionData, useNavigation } from "@remix-run/react";
import { login, register } from "~/actions/auth.server";
import { Mail, Lock, User, UserPlus, ArrowRight, AlertCircle, Eye, EyeOff, Loader2, Check } from "lucide-react";
import { useState } from "react";

// Simple loader - no auth check needed
export async function loader() {
  return json({});
}

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  const email = formData.get("email");
  const password = formData.get("password");
  const username = formData.get("username");
  const name = formData.get("name");

  // Validate form
  if (
    typeof email !== "string" ||
    typeof password !== "string" ||
    email.length === 0 ||
    password.length === 0
  ) {
    return json({ error: "Please provide both email and password" }, { status: 400 });
  }

  if (password.length < 6) {
    return json({ error: "Password must be at least 6 characters long" }, { status: 400 });
  }

  // Prepare registration data
  const registerData = {
    email,
    password,
    username: typeof username === "string" && username.length > 0 ? username : undefined,
    name: typeof name === "string" && name.length > 0 ? name : undefined
  };

  try {
    // Register the new user
    await register(registerData);
    
    // Log them in automatically
    const loginResult = await login({ 
      email: email as string, 
      password: password as string 
    });
    
    // Create redirect with cookie from login response
    const headers = new Headers({ 
      "Location": "/login?message=Registration successful. Please log in."
    });
    
    // Return redirect with proper cookie headers
    return new Response(null, {
      status: 302,
      headers
    });
  } catch (error) {
    console.error("Registration error:", error);
    return json({ 
      error: error instanceof Error ? error.message : "An unexpected error occurred" 
    }, { status: 400 });
  }
}

export default function Register() {
  const actionData = useActionData<typeof action>();
  const navigation = useNavigation();
  const isSubmitting = navigation.state === "submitting";
  const [showPassword, setShowPassword] = useState(false);
  const [password, setPassword] = useState("");

  // Password strength indicator
  const getPasswordStrength = (pass: string) => {
    if (pass.length === 0) return { strength: 0, label: "", color: "" };
    if (pass.length < 6) return { strength: 1, label: "Weak", color: "text-red-500" };
    if (pass.length < 8) return { strength: 2, label: "Fair", color: "text-yellow-500" };
    if (pass.length >= 8 && /[A-Z]/.test(pass) && /[0-9]/.test(pass)) {
      return { strength: 4, label: "Strong", color: "text-green-500" };
    }
    return { strength: 3, label: "Good", color: "text-blue-500" };
  };

  const passwordStrength = getPasswordStrength(password);

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-200 via-white to-indigo-300 dark:from-indigo-950 dark:via-gray-900 dark:to-violet-900 relative overflow-hidden">
      {/* Background decoration */}
      <div className="absolute inset-0 opacity-50 dark:opacity-70">
        <div className="absolute inset-0" style={{ backgroundImage: 'radial-gradient(#6366F1 1.5px, transparent 1.5px)', backgroundSize: '20px 20px' }}></div>
      </div>
      
      {/* Additional gradient overlay */}
      <div className="absolute inset-0 bg-gradient-to-r from-indigo-300/30 via-transparent to-purple-300/30 dark:from-indigo-600/30 dark:via-transparent dark:to-purple-600/30" />
      
      {/* Floating elements for visual interest */}
      <div className="absolute top-10 right-10 w-24 h-24 bg-purple-400/20 rounded-full blur-xl animate-pulse" />
      <div className="absolute bottom-32 left-10 w-28 h-28 bg-indigo-400/20 rounded-full blur-xl animate-pulse delay-1000" />
      <div className="absolute top-1/3 right-1/4 w-20 h-20 bg-violet-300/20 rounded-full blur-xl animate-pulse delay-500" />

      <div className="relative flex min-h-screen items-center justify-center px-4 py-12 sm:px-6 lg:px-8">
        <div className="w-full max-w-md space-y-8">
          {/* Header Section */}
          <div className="text-center space-y-6">
          
            
            {/* Title */}
            <div className="space-y-2">
              <h2 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
                Join Kubesa
              </h2>
              <p className="text-gray-600 dark:text-gray-300">
                Create your account to get started
              </p>
            </div>
          </div>

          {/* Error Messages */}
          {actionData?.error && (
            <div className="rounded-xl bg-red-50/80 dark:bg-red-900/20 border border-red-200/60 dark:border-red-800/30 p-4 backdrop-blur-sm">
              <div className="flex items-start">
                <div className="flex-shrink-0">
                  <AlertCircle className="h-4 w-4 text-red-500 mt-0.5" />
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-semibold text-red-800 dark:text-red-200">Registration Error</h3>
                  <p className="text-sm text-red-700 dark:text-red-300 mt-1">
                    {actionData.error}
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Register Form */}
          <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-xl p-8">
            <Form method="post" className="space-y-6">
              {/* Email Field */}
              <div className="space-y-2">
                <label htmlFor="email" className="block text-sm font-semibold text-gray-900 dark:text-white">
                  Email address <span className="text-red-500">*</span>
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <Mail className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    id="email"
                    name="email"
                    type="email"
                    autoComplete="email"
                    required
                    disabled={isSubmitting}
                    className="w-full pl-10 pr-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                    placeholder="Enter your email"
                  />
                </div>
              </div>

              {/* Username Field */}
              <div className="space-y-2">
                <label htmlFor="username" className="block text-sm font-semibold text-gray-900 dark:text-white">
                  Username <span className="text-xs font-normal text-gray-500 dark:text-gray-400">(Optional)</span>
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <User className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    id="username"
                    name="username"
                    type="text"
                    autoComplete="username"
                    disabled={isSubmitting}
                    className="w-full pl-10 pr-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                    placeholder="Choose a username"
                  />
                </div>
              </div>

              {/* Full Name Field */}
              <div className="space-y-2">
                <label htmlFor="name" className="block text-sm font-semibold text-gray-900 dark:text-white">
                  Full name <span className="text-xs font-normal text-gray-500 dark:text-gray-400">(Optional)</span>
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <User className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    id="name"
                    name="name"
                    type="text"
                    autoComplete="name"
                    disabled={isSubmitting}
                    className="w-full pl-10 pr-4 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                    placeholder="Enter your full name"
                  />
                </div>
              </div>

              {/* Password Field */}
              <div className="space-y-2">
                <label htmlFor="password" className="block text-sm font-semibold text-gray-900 dark:text-white">
                  Password <span className="text-red-500">*</span>
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <Lock className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    id="password"
                    name="password"
                    type={showPassword ? "text" : "password"}
                    autoComplete="new-password"
                    required
                    disabled={isSubmitting}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full pl-10 pr-12 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                    placeholder="Create a strong password"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    disabled={isSubmitting}
                    className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors disabled:opacity-50"
                  >
                    {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
                  </button>
                </div>
                
                {/* Password Strength Indicator */}
                {password.length > 0 && (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-xs">
                      <span className="text-gray-500 dark:text-gray-400">Password strength:</span>
                      <span className={`font-medium ${passwordStrength.color}`}>
                        {passwordStrength.label}
                      </span>
                    </div>
                    <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
                      <div 
                        className={`h-1.5 rounded-full transition-all duration-300 ${
                          passwordStrength.strength === 1 ? 'bg-red-500 w-1/4' :
                          passwordStrength.strength === 2 ? 'bg-yellow-500 w-2/4' :
                          passwordStrength.strength === 3 ? 'bg-blue-500 w-3/4' :
                          passwordStrength.strength === 4 ? 'bg-green-500 w-full' : 'w-0'
                        }`}
                      />
                    </div>
                  </div>
                )}
              </div>

              {/* Submit Button */}
              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full px-6 py-3 bg-gradient-to-r from-indigo-600 to-indigo-700 hover:from-indigo-700 hover:to-indigo-800 text-white font-semibold rounded-xl shadow-lg shadow-indigo-500/25 hover:shadow-indigo-500/40 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center transform hover:scale-105 disabled:hover:scale-100"
              >
                {isSubmitting ? (
                  <>
                    <Loader2 className="h-5 w-5 mr-2 animate-spin" />
                    Creating account...
                  </>
                ) : (
                  <>
                    Create account
                    <ArrowRight className="h-5 w-5 ml-2 transition-transform group-hover:translate-x-1" />
                  </>
                )}
              </button>
            </Form>
          </div>

          {/* Login Link */}
          <div className="text-center">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Already have an account?{" "}
              <Link
                to="/login"
                className="font-semibold text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 transition-colors"
              >
                Sign in here
              </Link>
            </p>
          </div>

          {/* Benefits Section */}
          <div className="bg-indigo-50/60 dark:bg-indigo-900/20 backdrop-blur-sm rounded-xl border border-indigo-200/50 dark:border-indigo-700/50 p-6">
            <h3 className="text-sm font-semibold text-indigo-800 dark:text-indigo-200 mb-3">🚀 What you'll get:</h3>
            <ul className="text-sm text-indigo-700 dark:text-indigo-300 space-y-2">
              <li className="flex items-center">
                <Check className="h-4 w-4 mr-2 text-indigo-600" />
                Deploy unlimited projects
              </li>
              <li className="flex items-center">
                <Check className="h-4 w-4 mr-2 text-indigo-600" />
                Real-time monitoring and logs
              </li>
              <li className="flex items-center">
                <Check className="h-4 w-4 mr-2 text-indigo-600" />
                Git-based deployment workflow
              </li>
            </ul>
          </div>

          {/* Additional Info */}
          <div className="text-center pt-4">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              By creating an account, you agree to our Terms of Service and Privacy Policy
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}