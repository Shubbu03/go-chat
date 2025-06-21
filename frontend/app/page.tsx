import { ThemeToggle } from "./(landing)/components/theme-toggle";
import { HeroSection } from "./(landing)/components/hero-section";
import { FeaturesSection } from "./(landing)/components/features-section";
import Footer from "@/components/footer";

export default function LandingPage() {
  return (
    <div className="min-h-screen">
      <ThemeToggle />
      <HeroSection />
      <FeaturesSection />
      <Footer />
    </div>
  );
}
