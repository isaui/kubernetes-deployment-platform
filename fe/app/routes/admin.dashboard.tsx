import { LoaderFunctionArgs, json } from '@remix-run/node';
import { useLoaderData } from '@remix-run/react';
import AdminDashboard from '~/components/dashboard/admin/AdminDashboard';
import {
  getClusterInfo,
  getNodeStats,
  getPodStats,
  getDeploymentStats,
  getServiceStats,
  getIngressStats,
  getCertificateStats,
  getPVCStats
} from '~/actions/stats.server';
import { getRegistries } from '~/actions/registry.server';

export async function loader({ request }: LoaderFunctionArgs) {
  try {
    // Fetch all required data in parallel
    const [
      clusterInfoData,
      nodesData,
      podsData,
      deploymentsData,
      servicesData,
      ingressesData,
      certificatesData,
      pvcStatsData,
      registriesData
    ] = await Promise.all([
      getClusterInfo(request),
      getNodeStats(request),
      getPodStats(request),
      getDeploymentStats(request),
      getServiceStats(request),
      getIngressStats(request),
      getCertificateStats(request),
      getPVCStats(request),
      getRegistries({}, request)
    ]);
    
    // Extract the actual data from the responses
    const clusterInfo = clusterInfoData;
    const { nodes } = nodesData;
    const { pods } = podsData;
    const { deployments } = deploymentsData;
    const { services } = servicesData;
    const { ingresses } = ingressesData;
    const { certificates } = certificatesData;
    const { pvcs } = pvcStatsData;
    const registries = registriesData.registries || [];

    return json({
      clusterInfo,
      nodes,
      pods,
      deployments,
      services,
      ingresses,
      registries,
      certificates,
      pvcs
    });
  } catch (error) {
    if (error instanceof Response) {
      throw error;
    }
    throw new Response('Failed to load dashboard data', { status: 500 });
  }
}

export default function AdminDashboardRoute() {
  const {
    clusterInfo,
    nodes,
    pods,
    deployments,
    services,
    ingresses,
    certificates,
    registries,
    pvcs
  } = useLoaderData<typeof loader>();

  return (
    <AdminDashboard
      clusterInfo={clusterInfo}
      nodes={nodes}
      pods={pods}
      deployments={deployments}
      services={services}
      ingresses={ingresses}
      certificates={certificates}
      registries={registries}
      pvcs={pvcs}
    />
  );
}
