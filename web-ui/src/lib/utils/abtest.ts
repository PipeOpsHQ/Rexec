/**
 * A/B Testing Utility for Rexec
 * 
 * Randomly assigns users to test variants and persists the assignment
 * so the same user sees the same variant on return visits.
 */

export type LandingVariant = 'original' | 'promo';

const AB_TEST_KEY = 'rexec_ab_landing';
const AB_TEST_VERSION = 'v1'; // Increment to reset all assignments
const ROTATION_HOURS = 5; // Rotate variant every 5 hours
const ROTATION_MS = ROTATION_HOURS * 60 * 60 * 1000;

interface ABTestAssignment {
    variant: LandingVariant;
    version: string;
    assignedAt: number;
}

/**
 * Get or assign a landing page variant for the current user
 * Rotates every 5 hours, prioritizes original (70%) over promo (30%)
 */
export function getLandingVariant(): LandingVariant {
    // Check for URL override (for testing)
    const params = new URLSearchParams(window.location.search);
    const override = params.get('landing');
    if (override === 'original' || override === 'promo') {
        return override;
    }

    const now = Date.now();

    // Check existing assignment
    try {
        const stored = localStorage.getItem(AB_TEST_KEY);
        if (stored) {
            const assignment: ABTestAssignment = JSON.parse(stored);
            // If same version AND not expired (5 hours), use stored variant
            const isExpired = (now - assignment.assignedAt) > ROTATION_MS;
            if (assignment.version === AB_TEST_VERSION && !isExpired) {
                return assignment.variant;
            }
        }
    } catch {
        // Ignore parse errors
    }

    // Assign new variant: 70% original, 30% promo (prioritize main landing page)
    const variant: LandingVariant = Math.random() < 0.7 ? 'original' : 'promo';
    
    const assignment: ABTestAssignment = {
        variant,
        version: AB_TEST_VERSION,
        assignedAt: now
    };

    try {
        localStorage.setItem(AB_TEST_KEY, JSON.stringify(assignment));
    } catch {
        // localStorage might be full or disabled
    }

    return variant;
}

/**
 * Track an A/B test event (for analytics)
 */
export function trackABEvent(event: string, properties?: Record<string, unknown>): void {
    const variant = getLandingVariant();
    
    // If PostHog is available, track the event
    if (typeof window !== 'undefined' && (window as unknown as { posthog?: { capture: (event: string, props: Record<string, unknown>) => void } }).posthog) {
        (window as unknown as { posthog: { capture: (event: string, props: Record<string, unknown>) => void } }).posthog.capture(event, {
            ab_test: 'landing_page',
            ab_variant: variant,
            ab_version: AB_TEST_VERSION,
            ...properties
        });
    }
    
    // Log to console in development
    if (import.meta.env.DEV) {
        console.log(`[A/B Test] ${event}`, { variant, ...properties });
    }
}

/**
 * Force a specific variant (for testing/debugging)
 */
export function forceVariant(variant: LandingVariant): void {
    const assignment: ABTestAssignment = {
        variant,
        version: AB_TEST_VERSION,
        assignedAt: Date.now()
    };
    localStorage.setItem(AB_TEST_KEY, JSON.stringify(assignment));
}

/**
 * Get time until next rotation (in minutes)
 */
export function getTimeUntilRotation(): number {
    try {
        const stored = localStorage.getItem(AB_TEST_KEY);
        if (stored) {
            const assignment: ABTestAssignment = JSON.parse(stored);
            const elapsed = Date.now() - assignment.assignedAt;
            const remaining = ROTATION_MS - elapsed;
            return Math.max(0, Math.ceil(remaining / (60 * 1000)));
        }
    } catch {
        // Ignore parse errors
    }
    return 0;
}

/**
 * Clear A/B test assignment (user will get a new random assignment)
 */
export function clearAssignment(): void {
    localStorage.removeItem(AB_TEST_KEY);
}
