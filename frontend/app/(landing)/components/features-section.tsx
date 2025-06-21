'use client';

import { useRef } from 'react';
import { motion, useInView } from 'framer-motion';
import {
    MessageCircle,
    Users,
    Shield,
    Zap,
    Smartphone,
    Globe,
} from 'lucide-react';

const FeatureCard = ({ icon: Icon, title, description, delay = 0 }: {
    icon: any;
    title: string;
    description: string;
    delay?: number;
}) => {
    const ref = useRef(null);
    const isInView = useInView(ref, { once: true, amount: 0.1 });

    return (
        <motion.div
            ref={ref}
            initial={{ opacity: 0, y: 50 }}
            animate={isInView ? { opacity: 1, y: 0 } : {}}
            className="group relative p-6 rounded-2xl bg-white/5 backdrop-blur-md border border-white/10 hover:border-cyan-400/30 transition-all duration-150 cursor-pointer"
            whileHover={{
                scale: 1.05,
                boxShadow: "0 20px 40px rgba(0, 173, 181, 0.1)"
            }}
        >
            <motion.div
                className="w-12 h-12 rounded-xl bg-gradient-to-r from-cyan-400 to-cyan-600 flex items-center justify-center mb-4 group-hover:shadow-lg transition-shadow"
                whileHover={{ rotate: 5 }}
            >
                <Icon className="w-6 h-6 text-white" />
            </motion.div>
            <h3 className="text-xl font-semibold mb-2 text-foreground">{title}</h3>
            <p className="text-foreground/70">{description}</p>
        </motion.div>
    );
};

export const FeaturesSection = () => {
    const features = [
        {
            icon: MessageCircle,
            title: "Real-time Messaging",
            description: "Instant message delivery with WebSocket connections for seamless communication."
        },
        {
            icon: Users,
            title: "Group Conversations",
            description: "Create and manage group chats with friends, family, or team members."
        },
        {
            icon: Shield,
            title: "End-to-End Security",
            description: "Your messages are encrypted and secure. Privacy is our top priority."
        },
        {
            icon: Zap,
            title: "Lightning Fast",
            description: "Optimized performance with smooth animations and instant responsiveness."
        },
        {
            icon: Smartphone,
            title: "Cross-Platform",
            description: "Available on all devices with responsive design and native-like experience."
        },
        {
            icon: Globe,
            title: "Global Reach",
            description: "Connect with people worldwide with multi-language support and localization."
        }
    ];

    return (
        <section className="py-20 px-6 bg-gradient-to-b from-background to-secondary/10">
            <div className="max-w-7xl mx-auto">
                <motion.div
                    className="text-center mb-16"
                    initial={{ opacity: 0, y: 30 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                >
                    <h2 className="text-4xl md:text-6xl font-bold mb-6 text-foreground">
                        Powerful <span className="text-gradient">Features</span>
                    </h2>
                    <p className="text-xl text-foreground/70 max-w-3xl mx-auto">
                        Everything you need for modern communication, wrapped in a beautiful and intuitive interface.
                    </p>
                </motion.div>

                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
                    {features.map((feature, index) => (
                        <FeatureCard
                            key={index}
                            icon={feature.icon}
                            title={feature.title}
                            description={feature.description}
                            delay={index * 0.1}
                        />
                    ))}
                </div>
            </div>
        </section>
    );
}; 