import { motion } from "framer-motion";

const GeometricBackground = () => {
  return (
    <div className="absolute inset-0 overflow-hidden pointer-events-none">
      <div className="absolute top-20 left-20 w-32 h-32 bg-gradient-to-br from-cyan-400/10 to-cyan-600/10 rounded-full blur-xl" />
      <div className="absolute top-40 right-32 w-24 h-24 bg-gradient-to-br from-cyan-500/15 to-cyan-700/15 rounded-lg rotate-45 blur-lg" />
      <div className="absolute bottom-32 left-1/4 w-20 h-20 bg-gradient-to-br from-cyan-300/20 to-cyan-500/20 rounded-full blur-sm" />
      <div className="absolute bottom-20 right-20 w-28 h-28 bg-gradient-to-br from-cyan-400/10 to-cyan-600/10 rounded-lg rotate-12 blur-md" />

      <div className="absolute inset-0 bg-gradient-to-br from-transparent via-cyan-400/5 to-transparent">
        <div
          className="absolute inset-0"
          style={{
            backgroundImage: `radial-gradient(circle at 1px 1px, rgba(0, 173, 181, 0.1) 1px, transparent 0)`,
            backgroundSize: "50px 50px",
          }}
        />
      </div>

      {[...Array(3)].map((_, i) => (
        <motion.div
          key={i}
          className="absolute w-2 h-2 bg-cyan-400/30 rounded-full"
          animate={{
            y: [0, -20, 0],
            opacity: [0.3, 0.7, 0.3],
          }}
          transition={{
            duration: 4 + i * 2,
            repeat: Infinity,
            ease: "easeInOut",
          }}
          style={{
            left: `${20 + i * 30}%`,
            top: `${20 + i * 20}%`,
          }}
        />
      ))}
    </div>
  );
};

export default GeometricBackground;
