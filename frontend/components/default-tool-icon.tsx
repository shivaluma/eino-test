import {
  BarChart3Icon,
  CodeIcon,
  GlobeIcon,
  HammerIcon,
  LineChartIcon,
  PieChartIcon,
  SearchIcon,
  TableIcon,
} from "lucide-react";
import { DefaultToolName } from "@/lib/ai/tools";

interface DefaultToolIconProps {
  name: DefaultToolName;
  className?: string;
}

export function DefaultToolIcon({ name, className }: DefaultToolIconProps) {
  switch (name) {
    case DefaultToolName.CreatePieChart:
      return <PieChartIcon className={className} />;
    case DefaultToolName.CreateBarChart:
      return <BarChart3Icon className={className} />;
    case DefaultToolName.CreateLineChart:
      return <LineChartIcon className={className} />;
    case DefaultToolName.CreateTable:
      return <TableIcon className={className} />;
    case DefaultToolName.WebSearch:
      return <SearchIcon className={className} />;
    case DefaultToolName.WebContent:
      return <GlobeIcon className={className} />;
    case DefaultToolName.Http:
      return <HammerIcon className={className} />;
    case DefaultToolName.JavascriptExecution:
    case DefaultToolName.PythonExecution:
      return <CodeIcon className={className} />;
    default:
      return <HammerIcon className={className} />;
  }
}