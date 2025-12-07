/* A simple rating algorithm */
import type {
  KUN_GALGAME_RATING_RECOMMEND_CONST,
  KUN_GALGAME_DIMENSIONS,
  KUN_GALGAME_RATING_PLAY_STATUS_CONST
} from '~/constants/galgame-rating'

interface DimsInput {
  art: number
  story: number
  music: number
  character: number
  route: number
  system: number
  voice: number
  replay_value: number
}

type KunGalgameRatingDim = (typeof KUN_GALGAME_DIMENSIONS)[number]
type KunGalgameRatingPlayStatus =
  (typeof KUN_GALGAME_RATING_PLAY_STATUS_CONST)[number]
type KunGalgameRatingRecommend =
  (typeof KUN_GALGAME_RATING_RECOMMEND_CONST)[number]

/**
 * Default dimension weights.
 * The sum should ideally be 1.0. The algorithm normalizes it if not.
 * These can be adjusted based on product requirements.
 */
const DEFAULT_DIM_WEIGHTS: Record<KunGalgameRatingDim, number> = {
  story: 0.25,
  character: 0.2,
  art: 0.15,
  route: 0.1,
  music: 0.08,
  voice: 0.06,
  system: 0.06,
  replay_value: 0.1
}

/**
 * Blending weight for the 'overall' score vs. the calculated dimension score.
 * `DEFAULT_OVERALL_WEIGHT` represents the proportion of the base score
 * that comes from the user's single 'overall' rating.
 */
const DEFAULT_OVERALL_WEIGHT = 0.4

/**
 * Adjustable: The maximum score adjustment (plus or minus) from the recommendation.
 * For example, 0.8 means the recommendation can alter the score by up to ±0.8 points.
 */
const RECOMMEND_INFLUENCE_POINTS = 0.8

/**
 * Maps the recommendation status to a numerical value [-1, 1].
 * This value is then scaled by `RECOMMEND_INFLUENCE_POINTS` to produce a score adjustment.
 */
const RECOMMEND_SCORE_MAP: Record<string, number> = {
  strong_no: -1,
  no: -0.5,
  neutral: 0,
  yes: 0.5,
  strong_yes: 1
}

/**
 * Direct score adjustment based on the user's play status.
 * This acts as a confidence modifier for the rating.
 */
const PLAY_STATUS_ADJUST: Record<KunGalgameRatingPlayStatus, number> = {
  not_started: -1.5,
  in_progress: -0.8,
  finished_one: -0.2,
  finished_main: 0,
  finished_all: 0.4, // A small bonus for full completion
  dropped: -1.0
}

export const calcGalgameRating = (
  dims: DimsInput,
  overall: number,
  play_status: KunGalgameRatingPlayStatus,
  recommend: KunGalgameRatingRecommend
): number => {
  // Helper functions
  const clamp = (v: number, lo: number, hi: number) =>
    Math.max(lo, Math.min(hi, v))
  const round1 = (v: number) => Math.round(v * 10) / 10

  // Sanitize and normalize inputs to ensure they are within the 0-10 range.
  const safeDims: Record<KunGalgameRatingDim, number> = {
    art: clamp(dims.art ?? 0, 0, 10),
    story: clamp(dims.story ?? 0, 0, 10),
    music: clamp(dims.music ?? 0, 0, 10),
    character: clamp(dims.character ?? 0, 0, 10),
    route: clamp(dims.route ?? 0, 0, 10),
    system: clamp(dims.system ?? 0, 0, 10),
    voice: clamp(dims.voice ?? 0, 0, 10),
    replay_value: clamp(dims.replay_value ?? 0, 0, 10)
  }
  const safeOverall = clamp(overall ?? 0, 0, 10)

  // Calculate the weighted average of dimension scores.
  // This includes normalization in case the weights don't sum to 1.
  const DIM_KEYS = Object.keys(DEFAULT_DIM_WEIGHTS) as KunGalgameRatingDim[]
  const weightSum =
    DIM_KEYS.reduce((s, k) => s + (DEFAULT_DIM_WEIGHTS[k] ?? 0), 0) || 1
  const dimsWeighted = DIM_KEYS.reduce(
    (s, k) => s + safeDims[k] * ((DEFAULT_DIM_WEIGHTS[k] ?? 0) / weightSum),
    0
  )

  // Blend the user's overall score with the calculated weighted dimension score.
  const overallWeight = clamp(DEFAULT_OVERALL_WEIGHT ?? 0.4, 0, 1)
  const dimsWeight = 1 - overallWeight
  const baseScore = safeOverall * overallWeight + dimsWeighted * dimsWeight // Result is 0..10

  // Calculate the score adjustment from the recommendation.
  const recVal = RECOMMEND_SCORE_MAP[recommend] ?? 0
  const recAdjust = recVal * RECOMMEND_INFLUENCE_POINTS

  // Get the direct score adjustment from the play status.
  const statusAdjust = PLAY_STATUS_ADJUST[play_status] ?? 0

  // Combine all parts and clamp the result to the final 0-10 range.
  let finalScore = baseScore + recAdjust + statusAdjust
  finalScore = clamp(finalScore, 0, 10)

  // Return the score rounded to one decimal place.
  return round1(finalScore)
}

// TODO:
const __TEST__ = () => {}
