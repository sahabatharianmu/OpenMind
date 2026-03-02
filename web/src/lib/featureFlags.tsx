import React, { createContext, useContext, useMemo } from "react";
import { withLDProvider, useFlags, useLDClient } from "launchdarkly-react-client-sdk";

const LD_CLIENT_ID = import.meta.env.VITE_LD_CLIENT_ID || "";

// Context for feature flag state
interface FeatureFlagContextType {
  isEnabled: (flagKey: string) => boolean;
  getVariation: <T = string>(flagKey: string, defaultValue: T) => T;
  isReady: boolean;
}

const FeatureFlagContext = createContext<FeatureFlagContextType>({
  isEnabled: () => false,
  getVariation: <T,>(_: string, defaultValue: T) => defaultValue,
  isReady: false,
});

/** Hook to check if a feature flag is enabled */
export function useFeatureFlag(flagKey: string): boolean {
  const flags = useFlags();
  return flags[flagKey] ?? false;
}

/** Hook to get a feature flag variation (for A/B testing) */
export function useFeatureFlagVariation<T = string>(
  flagKey: string,
  defaultValue: T
): T {
  const flags = useFlags();
  return (flags[flagKey] as T) ?? defaultValue;
}

/** Hook to access the full feature flag context */
export function useFeatureFlags(): FeatureFlagContextType {
  return useContext(FeatureFlagContext);
}

/** Inner provider that sets up the context */
function FeatureFlagInner({ children }: { children: React.ReactNode }) {
  const flags = useFlags();
  const ldClient = useLDClient();

  const value = useMemo<FeatureFlagContextType>(
    () => ({
      isEnabled: (flagKey: string) => flags[flagKey] ?? false,
      getVariation: <T,>(flagKey: string, defaultValue: T) =>
        (flags[flagKey] as T) ?? defaultValue,
      isReady: ldClient?.getContext() !== undefined,
    }),
    [flags, ldClient]
  );

  return (
    <FeatureFlagContext.Provider value={value}>
      {children}
    </FeatureFlagContext.Provider>
  );
}

/**
 * Wraps a component with the LaunchDarkly provider.
 * If no client ID is set, renders children directly (local/dev mode).
 */
export function withFeatureFlags<P extends object>(
  WrappedComponent: React.ComponentType<P>
): React.ComponentType<P> {
  if (!LD_CLIENT_ID) {
    return WrappedComponent;
  }

  const LDWrapped = withLDProvider({
    clientSideID: LD_CLIENT_ID,
    reactOptions: {
      useCamelCaseFlagKeys: true,
    },
    context: {
      kind: "user",
      key: "anonymous",
    },
  })(WrappedComponent as React.ComponentType);

  return LDWrapped as React.ComponentType<P>;
}

/** Provider component for use in App.tsx */
export function FeatureFlagProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  if (!LD_CLIENT_ID) {
    return <>{children}</>;
  }

  return <FeatureFlagInner>{children}</FeatureFlagInner>;
}
