"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z, ZodType } from "zod";
import { ArrowLeft, Eye, EyeOff, Mail, Lock, User } from "lucide-react";
import { useState } from "react";
import Link from "next/link";

interface FormField {
  name: string;
  label: string;
  type: string;
  placeholder: string;
  icon?: React.ReactNode;
}

interface AuthFormProps {
  title: string;
  subtitle: string;
  fields: FormField[];
  submitText: string;
  loadingText: string;
  onSubmit: (data: any) => void;
  bottomText: string;
  bottomLinkText: string;
  bottomLinkHref: string;
  imageUrl: string;
  imageAlt: string;
  validationSchema: ZodType<any, any, any>;
}

export default function AuthForm({
  title,
  subtitle,
  fields,
  submitText,
  loadingText,
  onSubmit,
  bottomText,
  bottomLinkText,
  bottomLinkHref,
  imageUrl,
  imageAlt,
  validationSchema,
}: AuthFormProps) {
  const [showPasswords, setShowPasswords] = useState<Record<string, boolean>>(
    {}
  );
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(validationSchema),
  });

  const handleFormSubmit = async (data: any) => {
    setIsLoading(true);
    try {
      onSubmit(data);
    } catch (error) {
      console.error("Form submission failed:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const togglePasswordVisibility = (fieldName: string) => {
    setShowPasswords((prev) => ({
      ...prev,
      [fieldName]: !prev[fieldName],
    }));
  };

  const getFieldIcon = (type: string) => {
    switch (type) {
      case "email":
        return <Mail className="w-5 h-5 text-primary/60" />;
      case "password":
        return <Lock className="w-5 h-5 text-primary/60" />;
      case "text":
        return <User className="w-5 h-5 text-primary/60" />;
      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen flex">
      <div className="flex-2/5 flex flex-col justify-center p-8 bg-background relative">
        <Link
          href="/"
          className="absolute top-4 left-4 text-white flex items-center gap-2 hover:underline"
        >
          <ArrowLeft className="w-4 h-4" />
          Back
        </Link>

        <div className="w-full max-w-md mx-auto space-y-8">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 bg-primary rounded-full mb-4">
              <span className="text-white font-bold text-2xl">💬</span>
            </div>
            <h1 className="text-2xl font-bold text-foreground mb-2">Go Chat</h1>
          </div>

          <div className="text-center space-y-2">
            <h2 className="text-3xl font-bold text-foreground">{title}</h2>
            <p className="text-gray-500">{subtitle}</p>
          </div>

          <form onSubmit={handleSubmit(handleFormSubmit)} className="space-y-6">
            <div className="space-y-4">
              {fields.map((field) => (
                <div key={field.name}>
                  <label
                    htmlFor={field.name}
                    className="block text-white font-medium mb-1"
                  >
                    {field.label}
                  </label>
                  <div className="relative">
                    <span className="absolute left-3 top-1/2 transform -translate-y-1/2 pointer-events-none">
                      {getFieldIcon(field.type)}
                    </span>
                    <input
                      id={field.name}
                      type={
                        field.type === "password"
                          ? showPasswords[field.name]
                            ? "text"
                            : "password"
                          : field.type
                      }
                      placeholder={field.placeholder}
                      className="w-full pl-11 pr-10 py-3 bg-muted border border-secondary/20 rounded-lg text-foreground placeholder-gray-400 focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-colors"
                      {...register(field.name)}
                    />
                    {field.type === "password" && (
                      <button
                        type="button"
                        onClick={() => togglePasswordVisibility(field.name)}
                        className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-500 hover:text-gray-500 transition-colors cursor-pointer"
                      >
                        {showPasswords[field.name] ? (
                          <EyeOff className="w-5 h-5" />
                        ) : (
                          <Eye className="w-5 h-5" />
                        )}
                      </button>
                    )}
                  </div>
                  {errors[field.name] && (
                    <p className="text-red-500 text-sm mt-1">
                      {(errors[field.name]?.message as string) ?? ""}
                    </p>
                  )}
                </div>
              ))}
            </div>

            <button
              type="submit"
              disabled={isLoading}
              className="w-full bg-primary hover:bg-primary/90 text-white font-semibold py-3 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            >
              {isLoading ? loadingText : submitText}
            </button>
          </form>

          <div className="text-center">
            <p className="text-gray-500">
              {bottomText}{" "}
              <a
                href={bottomLinkHref}
                className="text-primary hover:underline font-medium"
              >
                {bottomLinkText}
              </a>
            </p>
          </div>
        </div>
      </div>

      <div className="hidden md:flex flex-3/5 relative">
        <img
          src={imageUrl}
          alt={imageAlt}
          className="w-4xl h-full object-cover"
        />
      </div>
    </div>
  );
}
