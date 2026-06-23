import { ActionFunctionArgs, json, LoaderFunctionArgs } from "@remix-run/node";
import { Form, Link, useActionData, useNavigation, useLoaderData, useSearchParams } from "@remix-run/react";
import { login } from "~/actions/auth.server";
import { Mail, Lock, ArrowRight, Eye, EyeOff, Loader2 } from "lucide-react";
import { useState } from "react";

// Loader to get error messages from URL search parameters
export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  
  // Get error information from URL parameters
  const error = url.searchParams.get("error");
  const message = url.searchParams.get("message");
  const redirectTo = url.searchParams.get("redirectTo");

  console.log("LOADER Redirect to:", redirectTo);
  
  return json({ 
    error,
    message,
    redirectTo
  });
}

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  const email = formData.get("email");
  const password = formData.get("password");
  const redirectTo = formData.get("redirectTo") || "/projects";

  console.log("Redirect to:", redirectTo);

  // Validate form
  if (
    typeof email !== "string" ||
    typeof password !== "string" ||
    email.length === 0 ||
    password.length === 0
  ) {
    return json({ error: "Please provide both email and password" }, { status: 400 });
  }

  try {
    // Attempt login - get response with Set-Cookie header
    const result = await login({ email, password });
    
    // Use the redirectTo value or default to dashboard based on user role
    let redirectPath = typeof redirectTo === "string" ? redirectTo : "/projects";

    console.log("Redirect path:", redirectPath);
    
    // If user has administrator role and no specific redirectTo was provided via URL,
    // redirect to admin dashboard
    if (result.user.role === "admin" && redirectTo === "/projects") {
      redirectPath = "/admin/dashboard";
    }
    
    // Create redirect with cookie from login response
    const headers = new Headers({ 
      "Location": redirectPath
    });
    
    // Get the Set-Cookie header from the response
    const setCookie = result.response.headers.get("Set-Cookie");
    if (setCookie) {
      headers.append("Set-Cookie", setCookie);
    }
    
    // Return redirect with proper cookie headers
    return new Response(null, {
      status: 302,
      headers
    });
  } catch (error) {
    console.error("Login error:", error);
    return json({ 
      error: error instanceof Error ? error.message : "An unexpected error occurred" 
    }, { status: 401 });
  }
}

export default function Login() {
  const actionData = useActionData<typeof action>();
  const loaderData = useLoaderData<typeof loader>();
  const navigation = useNavigation();
  const isSubmitting = navigation.state === "submitting";
  const [searchParams] = useSearchParams();
  const [showPassword, setShowPassword] = useState(false);

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-200 via-white to-indigo-300 dark:from-indigo-950 dark:via-gray-900 dark:to-violet-900 relative overflow-hidden">
      {/* Background decoration */}
      <div className="absolute inset-0 opacity-50 dark:opacity-70">
        <div className="absolute inset-0" style={{ backgroundImage: 'radial-gradient(#6366F1 1.5px, transparent 1.5px)', backgroundSize: '20px 20px' }}></div>
      </div>
      
      {/* Additional gradient overlay */}
      <div className="absolute inset-0 bg-gradient-to-r from-indigo-300/30 via-transparent to-purple-300/30 dark:from-indigo-600/30 dark:via-transparent dark:to-purple-600/30" />
      
      {/* Floating elements for visual interest */}
      <div className="absolute top-20 left-10 w-20 h-20 bg-indigo-400/20 rounded-full blur-xl animate-pulse" />
      <div className="absolute bottom-20 right-10 w-32 h-32 bg-purple-400/20 rounded-full blur-xl animate-pulse delay-1000" />
      <div className="absolute top-1/2 left-1/4 w-16 h-16 bg-indigo-300/20 rounded-full blur-xl animate-pulse delay-500" />

      <div className="relative flex min-h-screen items-center justify-center px-4 py-12 sm:px-6 lg:px-8">
        <div className="w-full max-w-md space-y-8">
          {/* Header Section */}
          <div className="text-center space-y-6">            
            {/* Title */}
            <div className="space-y-2">
              <h2 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
                Welcome back
              </h2>
              <p className="text-gray-600 dark:text-gray-300">
                Sign in to continue to Kubesa
              </p>
            </div>
          </div>

          {/* Notification Messages */}
          {(actionData?.error || loaderData?.message) && (
            <div className="rounded-xl bg-amber-50/80 dark:bg-amber-900/20 border border-amber-200/60 dark:border-amber-800/30 p-4 backdrop-blur-sm">
              <div className="p-1">
                <p className="text-sm text-amber-700 dark:text-amber-300">
                  {actionData?.error || loaderData?.message}
                </p>
              </div>
            </div>
          )}

          {/* Login Form */}
          <div className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm rounded-2xl border border-gray-200/50 dark:border-gray-700/50 shadow-xl p-8">
            <Form method="post" className="space-y-6">
              {/* Hidden redirect field */}
              <input 
                type="hidden" 
                name="redirectTo" 
                value={loaderData.redirectTo || searchParams.get("redirectTo") || "/projects"} 
              />

              {/* Email Field */}
              <div className="space-y-2">
                <label htmlFor="email" className="block text-sm font-semibold text-gray-900 dark:text-white">
                  Email address
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

              {/* Password Field */}
              <div className="space-y-2">
                <label htmlFor="password" className="block text-sm font-semibold text-gray-900 dark:text-white">
                  Password
                </label>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <Lock className="h-5 w-5 text-gray-400" />
                  </div>
                  <input
                    id="password"
                    name="password"
                    type={showPassword ? "text" : "password"}
                    autoComplete="current-password"
                    required
                    disabled={isSubmitting}
                    className="w-full pl-10 pr-12 py-3 bg-gray-50/50 dark:bg-gray-900/50 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:border-indigo-500 dark:focus:border-indigo-400 focus:ring-2 focus:ring-indigo-500/20 transition-all backdrop-blur-sm disabled:opacity-50"
                    placeholder="Enter your password"
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
                    Signing in...
                  </>
                ) : (
                  <>
                    Sign in
                    <ArrowRight className="h-5 w-5 ml-2 transition-transform group-hover:translate-x-1" />
                  </>
                )}
              </button>
            </Form>
          </div>

          {/* Register Link */}
          <div className="text-center">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Don't have an account?{" "}
              <Link
                to="/register"
                className="font-semibold text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 transition-colors"
              >
                Create one here
              </Link>
            </p>
          </div>

          {/* Additional Info */}
          <div className="text-center pt-4">
            <p className="text-xs text-gray-500 dark:text-gray-400">
              By signing in, you agree to our Terms of Service and Privacy Policy
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}