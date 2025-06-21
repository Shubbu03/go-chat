"use client";

import { motion } from "framer-motion";
import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";

export const ThemeToggle = ({
  position = "fixed",
}: {
  position?: "fixed" | "static";
}) => {
  const { theme, setTheme } = useTheme();

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark");
  };

  const baseClasses =
    "p-3 rounded-full bg-white/10 backdrop-blur-md border border-white/20 text-foreground hover:bg-white/20 transition-colors cursor-pointer";
  const positionClasses =
    position === "fixed" ? "fixed top-6 right-6 z-50" : "";

  return (
    <motion.button
      onClick={toggleTheme}
      className={`${baseClasses} ${positionClasses}`}
      whileHover={{ scale: 1.1 }}
      whileTap={{ scale: 0.95 }}
    >
      {theme === "dark" ? (
        <Sun className="w-5 h-5 rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
      ) : (
        <Moon className="w-5 h-5 rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
      )}
    </motion.button>
  );
};
