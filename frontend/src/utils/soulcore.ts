export function getSoulCoreImageUrl(creatureName: string): string {
  const filename = `${creatureName.replaceAll(' ', '_')}_Soul_Core.gif`
  return `/assets/soulcores/${filename}`
}
