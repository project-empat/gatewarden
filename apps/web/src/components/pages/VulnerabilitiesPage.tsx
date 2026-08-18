import { useQuery } from "@tanstack/react-query";
import { Bug, FileWarning, ShieldAlert, PackageSearch } from "lucide-react";
import { api } from "@/api/client";

interface VulnerablePackage {
	node_id: string;
	hostname: string;
	name: string;
	version: string;
	cve_count: number;
	top_cve?: string;
	summary?: string;
	checked_at: string;
}

interface FIMChange {
	node_id: string;
	hostname: string;
	path: string;
	changed_at?: string;
}

interface SecuritySummary {
	vulnerable_packages: number;
	security_updates_pending: number;
	fim_changes: number;
}

function StatCard({
	title,
	value,
	icon: Icon,
	tone,
}: {
	title: string;
	value: number;
	icon: typeof Bug;
	tone: string;
}) {
	return (
		<div className="bg-base-100 rounded-xl p-4 border border-base-content/5">
			<div className="flex items-center justify-between">
				<p className="text-xs text-base-content/40">{title}</p>
				<Icon className={`w-4 h-4 ${tone}`} />
			</div>
			<p className={`text-2xl font-bold mt-1 ${value > 0 ? tone : ""}`}>
				{value}
			</p>
		</div>
	);
}

export function VulnerabilitiesPage() {
	const { data: summary } = useQuery<SecuritySummary>({
		queryKey: ["security-summary"],
		queryFn: () => api.get("api/dashboard/security-summary").json(),
		refetchInterval: 30_000,
	});

	const { data: vulns } = useQuery<VulnerablePackage[]>({
		queryKey: ["vulnerabilities"],
		queryFn: () => api.get("api/vulnerabilities").json(),
		refetchInterval: 60_000,
	});

	const { data: fim } = useQuery<FIMChange[]>({
		queryKey: ["fim-changes"],
		queryFn: () => api.get("api/fim").json(),
		refetchInterval: 60_000,
	});

	return (
		<div className="space-y-6">
			<div>
				<h1 className="text-2xl font-bold">Vulnerabilities</h1>
				<p className="text-base-content/60 text-sm mt-1">
					Packages with known CVEs, pending security updates, and modified
					critical files
				</p>
			</div>

			<div className="grid grid-cols-1 md:grid-cols-3 gap-4">
				<StatCard
					title="Packages with Known CVEs"
					value={summary?.vulnerable_packages ?? vulns?.length ?? 0}
					icon={Bug}
					tone="text-error"
				/>
				<StatCard
					title="Security Updates Pending"
					value={summary?.security_updates_pending ?? 0}
					icon={PackageSearch}
					tone="text-warning"
				/>
				<StatCard
					title="Critical Files Changed"
					value={summary?.fim_changes ?? fim?.length ?? 0}
					icon={FileWarning}
					tone="text-warning"
				/>
			</div>

			{/* Vulnerable packages */}
			<div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
				<div className="px-5 py-4 border-b border-base-content/5 flex items-center gap-2">
					<ShieldAlert className="w-5 h-5 text-base-content/40" />
					<h2 className="font-semibold">Known CVEs</h2>
				</div>
				{!vulns || vulns.length === 0 ? (
					<div className="p-6 text-center">
						<Bug className="w-8 h-8 mx-auto text-base-content/20 mb-2" />
						<p className="text-sm text-base-content/50">
							No packages with known CVEs found.
						</p>
					</div>
				) : (
					<div className="divide-y divide-base-content/5">
						{vulns.map((v) => (
							<div
								key={`${v.node_id}-${v.name}-${v.version}`}
								className="px-5 py-3 flex items-center justify-between gap-4"
							>
								<div className="min-w-0">
									<p className="text-sm font-medium">
										<span className="font-mono">{v.name}</span>{" "}
										<span className="text-base-content/50 font-mono text-xs">
											{v.version}
										</span>
									</p>
									<p className="text-xs text-base-content/40 truncate">
										{v.hostname} &middot; {v.top_cve || `${v.cve_count} CVEs`}
										{v.summary ? ` — ${v.summary}` : ""}
									</p>
								</div>
								<span className="badge badge-error badge-sm shrink-0">
									{v.cve_count} CVE
								</span>
							</div>
						))}
					</div>
				)}
			</div>

			{/* FIM changes */}
			<div className="bg-base-100 rounded-xl border border-base-content/5 overflow-hidden">
				<div className="px-5 py-4 border-b border-base-content/5 flex items-center gap-2">
					<FileWarning className="w-5 h-5 text-base-content/40" />
					<h2 className="font-semibold">Modified Critical Files</h2>
				</div>
				{!fim || fim.length === 0 ? (
					<div className="p-6 text-center">
						<FileWarning className="w-8 h-8 mx-auto text-base-content/20 mb-2" />
						<p className="text-sm text-base-content/50">
							No monitored-file changes detected.
						</p>
					</div>
				) : (
					<div className="divide-y divide-base-content/5">
						{fim.map((f) => (
							<div
								key={`${f.node_id}-${f.path}`}
								className="px-5 py-3 flex items-center justify-between gap-4"
							>
								<div className="min-w-0">
									<p className="text-sm font-mono truncate">{f.path}</p>
									<p className="text-xs text-base-content/40">{f.hostname}</p>
								</div>
								<span className="text-xs text-warning shrink-0">
									{f.changed_at
										? new Date(f.changed_at).toLocaleString()
										: "changed"}
								</span>
							</div>
						))}
					</div>
				)}
			</div>
		</div>
	);
}
