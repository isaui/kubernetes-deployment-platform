import { ActionFunctionArgs, json, redirect } from '@remix-run/node';
import { Form, useActionData, useNavigation } from '@remix-run/react';
import { createRegistry } from '~/actions/registry.server';

type ActionErrors = {
  name: string | null;
  form?: string | null;
};

type ActionData = {
  errors?: ActionErrors;
  values?: {
    name?: string;
    isDefault?: boolean;
  };
};

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  
  const name = formData.get('name') as string;
  const isDefault = formData.get('isDefault') === 'true';
  
  // Validasi input
  const errors: ActionErrors = {
    name: name ? null : 'Nama registry harus diisi',
  };
  
  // Check if there are any errors
  if (Object.values(errors).some(error => error !== null)) {
    return json<ActionData>({ errors, values: { name, isDefault } });
  }
  
  try {
    // Panggil service untuk membuat registry
    await createRegistry({
      name,
      isDefault
    }, request);
    
    // Redirect ke halaman admin registries jika berhasil
    return redirect(`/admin/dashboard`);
  } catch (error: any) {
    return json<ActionData>({ 
      errors: { name: null, form: error.message || 'Gagal membuat registry' },
      values: { name, isDefault }
    });
  }
}

export default function NewRegistry() {
  const actionData = useActionData<ActionData>();
  const navigation = useNavigation();
  const isSubmitting = navigation.state === 'submitting';
  
  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
      <div className="md:flex md:items-center md:justify-between mb-6">
        <div className="flex-1 min-w-0">
          <h2 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
            Tambah Registry Baru
          </h2>
        </div>
        <div className="mt-4 flex md:mt-0 md:ml-4">
          <a
            href="/admin/dashboard"
            className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            Batal
          </a>
        </div>
      </div>

      <div className="bg-white shadow overflow-hidden sm:rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <Form method="post" className="space-y-6">
            {actionData?.errors?.form && (
              <div className="rounded-md bg-red-50 p-4">
                <div className="flex">
                  <div className="flex-shrink-0">
                    <svg className="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                    </svg>
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-medium text-red-800">{actionData.errors.form}</h3>
                  </div>
                </div>
              </div>
            )}

            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                Nama Registry
              </label>
              <div className="mt-1">
                <input
                  type="text"
                  name="name"
                  id="name"
                  defaultValue={actionData?.values?.name}
                  className={`shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md ${
                    actionData?.errors?.name ? 'border-red-300' : ''
                  }`}
                  placeholder="registry-production"
                />
                {actionData?.errors?.name && (
                  <p className="mt-2 text-sm text-red-600">{actionData.errors.name}</p>
                )}
              </div>
              <p className="mt-2 text-sm text-gray-500">
                Nama unik untuk registry ini
              </p>
            </div>
            
            <div className="relative flex items-start">
              <div className="flex items-center h-5">
                <input
                  id="isDefault"
                  name="isDefault"
                  type="checkbox"
                  value="true"
                  defaultChecked={actionData?.values?.isDefault}
                  className="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
                />
              </div>
              <div className="ml-3 text-sm">
                <label htmlFor="isDefault" className="font-medium text-gray-700">
                  Jadikan Default Registry
                </label>
                <p className="text-gray-500">
                  Registry ini akan digunakan secara default untuk semua deployment
                </p>
              </div>
            </div>
            
            <div className="pt-5">
              <div className="flex justify-end">
                <a
                  href="/admin/dashboard"
                  className="bg-white py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Batal
                </a>
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className={`ml-3 inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 ${
                    isSubmitting ? 'opacity-75 cursor-not-allowed' : ''
                  }`}
                >
                  {isSubmitting ? 'Menyimpan...' : 'Simpan'}
                </button>
              </div>
            </div>
          </Form>
        </div>
      </div>
    </div>
  );
}
