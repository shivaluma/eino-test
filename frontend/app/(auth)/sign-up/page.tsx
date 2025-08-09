"use client";

import { useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useObjectState } from "@/hooks/use-object-state";
import { cn } from "@/lib/utils";
import { ChevronLeft, Loader } from "lucide-react";
import { toast } from "sonner";
import { useCheckEmailMutation, useRegisterMutation } from "@/lib/mutations/auth";
import { useTranslations } from "next-intl";

export default function SignUpPage() {
  const t = useTranslations();
  const [step, setStep] = useState(1);
  const checkEmailMutation = useCheckEmailMutation();
  const registerMutation = useRegisterMutation();
  
  const [formData, setFormData] = useObjectState({
    email: "",
    name: "",
    password: "",
  });

  const steps = [
    t("Auth.SignUp.step1"),
    t("Auth.SignUp.step2"),
    t("Auth.SignUp.step3"),
  ];

  const backStep = () => {
    setStep(Math.max(step - 1, 1));
  };

  const successEmailStep = async () => {
    // Basic email validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(formData.email)) {
      toast.error(t("Auth.SignUp.invalidEmail"));
      return;
    }
    
    try {
      const result = await checkEmailMutation.mutateAsync({ email: formData.email });
      if (result.exists) {
        toast.error(t("Auth.SignUp.emailAlreadyExists"));
        return;
      }
      setStep(2);
    } catch (_) {
      toast.error("Failed to check email availability");
    }
  };

  const successNameStep = () => {
    if (!formData.name || formData.name.trim().length === 0) {
      toast.error(t("Auth.SignUp.nameRequired"));
      return;
    }
    setStep(3);
  };

  const successPasswordStep = async () => {
    if (!formData.password || formData.password.length < 8) {
      toast.error(t("Auth.SignUp.passwordRequired"));
      return;
    }
    
    await registerMutation.mutateAsync({
      email: formData.email,
      name: formData.name,
      password: formData.password,
    });
  };

  const isLoading = checkEmailMutation.isPending || registerMutation.isPending;

  return (
    <div className="animate-in fade-in duration-1000 w-full h-full flex flex-col p-4 md:p-8 justify-center relative">
      <div className="w-full flex justify-end absolute top-0 right-0">
        <Link href="/sign-in">
          <Button variant="ghost">{t("Auth.SignUp.signIn")}</Button>
        </Link>
      </div>
      <Card className="w-full md:max-w-md bg-background border-none mx-auto gap-0 shadow-none">
        <CardHeader>
          <CardTitle className="text-2xl text-center ">
            {t("Auth.SignUp.title")}
          </CardTitle>
          <CardDescription className="py-12">
            <div className="flex flex-col gap-2">
              <p className="text-xs text-muted-foreground text-right">
                Step {step} of {steps.length}
              </p>
              <div className="h-2 w-full relative bg-input">
                <div
                  style={{
                    width: `${(step / 3) * 100}%`,
                  }}
                  className="h-full bg-primary transition-all duration-300"
                ></div>
              </div>
            </div>
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-2">
            {step === 1 && (
              <div className={cn("flex flex-col gap-2")}>
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="mcp@example.com"
                  disabled={isLoading}
                  autoFocus
                  value={formData.email}
                  onKeyDown={(e) => {
                    if (
                      e.key === "Enter" &&
                      e.nativeEvent.isComposing === false
                    ) {
                      successEmailStep();
                    }
                  }}
                  onChange={(e) => setFormData({ email: e.target.value })}
                  required
                />
              </div>
            )}
            {step === 2 && (
              <div className={cn("flex flex-col gap-2")}>
                <Label htmlFor="name">Full Name</Label>
                <Input
                  id="name"
                  type="text"
                  placeholder="John Doe"
                  disabled={isLoading}
                  autoFocus
                  value={formData.name}
                  onKeyDown={(e) => {
                    if (
                      e.key === "Enter" &&
                      e.nativeEvent.isComposing === false
                    ) {
                      successNameStep();
                    }
                  }}
                  onChange={(e) => setFormData({ name: e.target.value })}
                  required
                />
              </div>
            )}
            {step === 3 && (
              <div className={cn("flex flex-col gap-2")}>
                <div className="flex items-center">
                  <Label htmlFor="password">Password</Label>
                </div>
                <Input
                  id="password"
                  type="password"
                  placeholder="********"
                  disabled={isLoading}
                  autoFocus
                  value={formData.password}
                  onKeyDown={(e) => {
                    if (
                      e.key === "Enter" &&
                      e.nativeEvent.isComposing === false
                    ) {
                      successPasswordStep();
                    }
                  }}
                  onChange={(e) => setFormData({ password: e.target.value })}
                  required
                />
              </div>
            )}
            <p className="text-muted-foreground text-xs mb-6">
              {steps[step - 1]}
            </p>
            <div className="flex gap-2">
              <Button
                disabled={isLoading}
                className={cn(step === 1 && "opacity-0", "w-1/2")}
                variant="ghost"
                onClick={backStep}
              >
                <ChevronLeft className="size-4" />
                {t("Common.back")}
              </Button>
              <Button
                disabled={isLoading}
                className="w-1/2"
                onClick={() => {
                  if (step === 1) successEmailStep();
                  if (step === 2) successNameStep();
                  if (step === 3) successPasswordStep();
                }}
              >
                {step === 3 ? t("Auth.SignUp.createAccount") : t("Common.next")}
                {isLoading && <Loader className="size-4 ml-2 animate-spin" />}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}