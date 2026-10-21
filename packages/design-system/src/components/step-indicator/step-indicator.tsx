import styles from './step-indicator.module.css';

export interface StepIndicatorProps {
  // Names the list, for example "Signing progress".
  label: string;
  steps: string[];
  // Zero-based index of the current step.
  current: number;
  // Read after each completed step, for example "completed", so completion is not shown by style only.
  completedLabel: string;
}

export function StepIndicator({ label, steps, current, completedLabel }: StepIndicatorProps) {
  return (
    <ol className={styles.steps} aria-label={label}>
      {steps.map((step, index) => {
        const state = index < current ? 'complete' : index === current ? 'current' : 'upcoming';
        return (
          <li
            key={step}
            className={`${styles.step} ${styles[state]}`}
            aria-current={state === 'current' ? 'step' : undefined}
          >
            <span className={styles.marker} aria-hidden="true">
              {index + 1}
            </span>
            <span className={styles.label}>
              {step}
              {state === 'complete' && <span className={styles.hidden}> {completedLabel}</span>}
            </span>
          </li>
        );
      })}
    </ol>
  );
}
