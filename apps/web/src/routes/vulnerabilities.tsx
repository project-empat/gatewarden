import { createRoute } from "@tanstack/react-router";
import { Route as rootRoute } from "./__root";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { VulnerabilitiesPage } from "@/components/pages/VulnerabilitiesPage";

export const Route = createRoute({
	getParentRoute: () => rootRoute,
	path: "/vulnerabilities",
	component: () => (
		<DashboardLayout>
			<VulnerabilitiesPage />
		</DashboardLayout>
	),
});
