import { 
  Shield, 
  Users, 
  FileText, 
  Calendar, 
  Receipt, 
  Eye,
  Database,
  Key
} from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

const features = [
  {
    icon: Shield,
    title: "End-to-End Encryption",
    description: "Clinical notes are encrypted before hitting the database. Even admins can't read them.",
  },
  {
    icon: Users,
    title: "Patient Management",
    description: "Comprehensive patient profiles with secure search and HIPAA-compliant data handling.",
  },
  {
    icon: FileText,
    title: "SOAP Notes",
    description: "Structured clinical documentation with note locking and immutable amendment history.",
  },
  {
    icon: Calendar,
    title: "Appointment Scheduling",
    description: "Intuitive calendar with conflict detection, reminders, and client self-booking.",
  },
  {
    icon: Receipt,
    title: "Billing & Superbills",
    description: "Generate professional invoices and insurance superbills with one click.",
  },
  {
    icon: Eye,
    title: "Audit Logging",
    description: "Immutable access logs track every patient record view for compliance.",
  },
  {
    icon: Database,
    title: "Data Sovereignty",
    description: "Self-host on your own infrastructure. Export your data anytime.",
  },
  {
    icon: Key,
    title: "Role-Based Access",
    description: "Granular permissions for clinicians, admins, and case managers.",
  },
];

const Features = () => {
  return (
    <section className="py-24 bg-card">
      <div className="container px-4">
        <div className="text-center mb-16">
          <h2 className="text-3xl md:text-4xl font-bold mb-4">
            Everything You Need to Run Your Practice
          </h2>
          <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
            Built by clinicians, for clinicians. Every feature designed with privacy and simplicity in mind.
          </p>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {features.map((feature, index) => (
            <Card 
              key={index} 
              className="group hover:shadow-lg transition-all duration-300 hover:-translate-y-1 border-border/50 bg-background"
            >
              <CardHeader>
                <div className="w-12 h-12 rounded-lg bg-accent flex items-center justify-center mb-4 group-hover:bg-primary/10 transition-colors">
                  <feature.icon className="w-6 h-6 text-primary" />
                </div>
                <CardTitle className="text-lg">{feature.title}</CardTitle>
              </CardHeader>
              <CardContent>
                <CardDescription className="text-muted-foreground">
                  {feature.description}
                </CardDescription>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
};

export default Features;
